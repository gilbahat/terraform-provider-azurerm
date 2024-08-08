// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicebus

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/go-azure-helpers/lang/response"
	"github.com/hashicorp/go-azure-sdk/resource-manager/servicebus/2021-06-01-preview/topics"
	"github.com/hashicorp/go-azure-sdk/resource-manager/servicebus/2022-10-01-preview/namespaces"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/validate"
	"github.com/hashicorp/terraform-provider-azurerm/internal/clients"
	"github.com/hashicorp/terraform-provider-azurerm/internal/features"
	azValidate "github.com/hashicorp/terraform-provider-azurerm/internal/services/servicebus/validate"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/internal/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceServiceBusTopic() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceServiceBusTopicCreateUpdate,
		Read:   resourceServiceBusTopicRead,
		Update: resourceServiceBusTopicCreateUpdate,
		Delete: resourceServiceBusTopicDelete,

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := topics.ParseTopicID(id)
			return err
		}),

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		Schema: resourceServiceBusTopicSchema(),
	}
}

func resourceServiceBusTopicSchema() map[string]*pluginsdk.Schema {
	schema := map[string]*pluginsdk.Schema{
		"name": {
			Type:         pluginsdk.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: azValidate.TopicName(),
		},

		//lintignore: S013
		"namespace_id": {
			Type:         pluginsdk.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: namespaces.ValidateNamespaceID,
		},

		"status": {
			Type:     pluginsdk.TypeString,
			Optional: true,
			Default:  string(topics.EntityStatusActive),
			ValidateFunc: validation.StringInSlice([]string{
				string(topics.EntityStatusActive),
				string(topics.EntityStatusDisabled),
			}, false),
		},

		"auto_delete_on_idle": {
			Type:         pluginsdk.TypeString,
			Optional:     true,
			Default:      "P10675199DT2H48M5.4775807S", // Never
			ValidateFunc: validate.ISO8601Duration,
		},

		"default_message_ttl": {
			Type:         pluginsdk.TypeString,
			Optional:     true,
			Default:      "P10675199DT2H48M5.4775807S", // Unbounded
			ValidateFunc: validate.ISO8601Duration,
		},

		"duplicate_detection_history_time_window": {
			Type:         pluginsdk.TypeString,
			Optional:     true,
			Default:      "PT10M", // 10 minutes
			ValidateFunc: validate.ISO8601Duration,
		},

		"batched_operations_enabled": {
			Type:     pluginsdk.TypeBool,
			Computed: !features.FourPointOhBeta(),
			Optional: true,
		},

		"express_enabled": {
			Type:     pluginsdk.TypeBool,
			Computed: !features.FourPointOhBeta(),
			Optional: true,
		},

		"partitioning_enabled": {
			Type:     pluginsdk.TypeBool,
			Computed: !features.FourPointOhBeta(),
			Optional: true,
			ForceNew: true,
		},

		"max_message_size_in_kilobytes": {
			Type:     pluginsdk.TypeInt,
			Optional: true,
			// NOTE: O+C this gets a variable default based on the sku and can be updated without issues
			Computed:     true,
			ValidateFunc: azValidate.ServiceBusMaxMessageSizeInKilobytes(),
		},

		"max_size_in_megabytes": {
			Type:     pluginsdk.TypeInt,
			Optional: true,
			// NOTE: O+C this gets a variable default based on the sku and can be updated without issues
			Computed:     true,
			ValidateFunc: azValidate.ServiceBusMaxSizeInMegabytes(),
		},

		"requires_duplicate_detection": {
			Type:     pluginsdk.TypeBool,
			Optional: true,
			ForceNew: true,
		},

		"support_ordering": {
			Type:     pluginsdk.TypeBool,
			Optional: true,
		},
	}

	if !features.FourPointOhBeta() {
		schema["auto_delete_on_idle"] = &pluginsdk.Schema{
			Type:         pluginsdk.TypeString,
			Optional:     true,
			Computed:     true,
			ValidateFunc: validate.ISO8601Duration,
		}

		schema["default_message_ttl"] = &pluginsdk.Schema{
			Type:         pluginsdk.TypeString,
			Optional:     true,
			Computed:     true,
			ValidateFunc: validate.ISO8601Duration,
		}

		schema["duplicate_detection_history_time_window"] = &pluginsdk.Schema{
			Type:         pluginsdk.TypeString,
			Optional:     true,
			Computed:     true,
			ValidateFunc: validate.ISO8601Duration,
		}

		schema["enable_batched_operations"] = &pluginsdk.Schema{
			Type:          pluginsdk.TypeBool,
			Optional:      true,
			Computed:      true,
			ConflictsWith: []string{"batched_operations_enabled"},
			Deprecated:    "The property `enable_batched_operations` has been superseded by `batched_operations_enabled` and will be removed in v4.0 of the AzureRM Provider.",
		}

		schema["enable_express"] = &pluginsdk.Schema{
			Type:          pluginsdk.TypeBool,
			Optional:      true,
			Computed:      true,
			ConflictsWith: []string{"express_enabled"},
			Deprecated:    "The property `enable_express` has been superseded by `express_enabled` and will be removed in v4.0 of the AzureRM Provider.",
		}

		schema["enable_partitioning"] = &pluginsdk.Schema{
			Type:          pluginsdk.TypeBool,
			Optional:      true,
			ForceNew:      true,
			Computed:      true,
			ConflictsWith: []string{"partitioning_enabled"},
			Deprecated:    "The property `enable_partitioning` has been superseded by `partitioning_enabled` and will be removed in v4.0 of the AzureRM Provider.",
		}
	}

	return schema
}

func resourceServiceBusTopicCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ServiceBus.TopicsClient
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()
	log.Printf("[INFO] preparing arguments for Azure ServiceBus Topic creation.")

	var id topics.TopicId
	if namespaceIdLit := d.Get("namespace_id").(string); namespaceIdLit != "" {
		namespaceId, err := topics.ParseNamespaceID(namespaceIdLit)
		if err != nil {
			return err
		}
		id = topics.NewTopicID(namespaceId.SubscriptionId, namespaceId.ResourceGroupName, namespaceId.NamespaceName, d.Get("name").(string))
	}

	if d.IsNewResource() {
		existing, err := client.Get(ctx, id)
		if err != nil {
			if !response.WasNotFound(existing.HttpResponse) {
				return fmt.Errorf("checking for presence of existing %s: %+v", id, err)
			}
		}

		if !response.WasNotFound(existing.HttpResponse) {
			return tf.ImportAsExistsError("azurerm_servicebus_topic", id.ID())
		}
	}

	enableBatchedOperations := d.Get("batched_operations_enabled").(bool)
	enableExpress := d.Get("express_enabled").(bool)
	enablePartitioning := d.Get("partitioning_enabled").(bool)
	if !features.FourPointOh() {
		if v := d.GetRawConfig().AsValueMap()["enable_batched_operations"]; !v.IsNull() {
			enableBatchedOperations = d.Get("enable_batched_operations").(bool)
		}

		if v := d.GetRawConfig().AsValueMap()["enable_express"]; !v.IsNull() {
			enableExpress = d.Get("enable_express").(bool)
		}

		if v := d.GetRawConfig().AsValueMap()["enable_partitioning"]; !v.IsNull() {
			enablePartitioning = d.Get("enable_partitioning").(bool)
		}
	}

	status := topics.EntityStatus(d.Get("status").(string))
	parameters := topics.SBTopic{
		Name: utils.String(id.TopicName),
		Properties: &topics.SBTopicProperties{
			Status:                     &status,
			EnableBatchedOperations:    utils.Bool(enableBatchedOperations),
			EnableExpress:              utils.Bool(enableExpress),
			EnablePartitioning:         utils.Bool(enablePartitioning),
			MaxSizeInMegabytes:         utils.Int64(int64(d.Get("max_size_in_megabytes").(int))),
			RequiresDuplicateDetection: utils.Bool(d.Get("requires_duplicate_detection").(bool)),
			SupportOrdering:            utils.Bool(d.Get("support_ordering").(bool)),
		},
	}

	if autoDeleteOnIdle := d.Get("auto_delete_on_idle").(string); autoDeleteOnIdle != "" {
		parameters.Properties.AutoDeleteOnIdle = utils.String(autoDeleteOnIdle)
	}

	if defaultTTL := d.Get("default_message_ttl").(string); defaultTTL != "" {
		parameters.Properties.DefaultMessageTimeToLive = utils.String(defaultTTL)
	}

	if duplicateWindow := d.Get("duplicate_detection_history_time_window").(string); duplicateWindow != "" {
		parameters.Properties.DuplicateDetectionHistoryTimeWindow = utils.String(duplicateWindow)
	}

	// We need to retrieve the namespace because Premium namespace works differently from Basic and Standard
	namespacesClient := meta.(*clients.Client).ServiceBus.NamespacesClient
	namespaceId := namespaces.NewNamespaceID(id.SubscriptionId, id.ResourceGroupName, id.NamespaceName)
	resp, err := namespacesClient.Get(ctx, namespaceId)
	if err != nil {
		return fmt.Errorf("retrieving ServiceBus Namespace %q (Resource Group %q): %+v", id.NamespaceName, id.ResourceGroupName, err)
	}

	// output of `max_message_size_in_kilobytes` is also set in non-Premium namespaces, with a value of 256
	if v, ok := d.GetOk("max_message_size_in_kilobytes"); ok && v.(int) != 256 {
		if model := resp.Model; model != nil {
			if model.Sku.Name != namespaces.SkuNamePremium {
				return fmt.Errorf("%s does not support input on `max_message_size_in_kilobytes` in %s SKU and should be removed", id, model.Sku.Name)
			}
			parameters.Properties.MaxMessageSizeInKilobytes = utils.Int64(int64(v.(int)))
		}
	}

	if _, err := client.CreateOrUpdate(ctx, id, parameters); err != nil {
		return fmt.Errorf("creating/updating %s: %v", id, err)
	}

	d.SetId(id.ID())
	return resourceServiceBusTopicRead(d, meta)
}

func resourceServiceBusTopicRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ServiceBus.TopicsClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := topics.ParseTopicID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, *id)
	if err != nil {
		if response.WasNotFound(resp.HttpResponse) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("retrieving %s: %+v", id, err)
	}

	d.Set("name", id.TopicName)
	d.Set("namespace_id", topics.NewNamespaceID(id.SubscriptionId, id.ResourceGroupName, id.NamespaceName).ID())

	if model := resp.Model; model != nil {
		if props := model.Properties; props != nil {
			status := ""
			if v := props.Status; v != nil {
				status = string(*v)
			}
			d.Set("status", status)
			d.Set("auto_delete_on_idle", props.AutoDeleteOnIdle)
			d.Set("default_message_ttl", props.DefaultMessageTimeToLive)

			if window := props.DuplicateDetectionHistoryTimeWindow; window != nil && *window != "" {
				d.Set("duplicate_detection_history_time_window", window)
			}

			if !features.FourPointOhBeta() {
				d.Set("enable_batched_operations", props.EnableBatchedOperations)
				d.Set("enable_express", props.EnableExpress)
				d.Set("enable_partitioning", props.EnablePartitioning)
			}

			d.Set("batched_operations_enabled", props.EnableBatchedOperations)
			d.Set("express_enabled", props.EnableExpress)
			d.Set("partitioning_enabled", props.EnablePartitioning)

			d.Set("max_message_size_in_kilobytes", props.MaxMessageSizeInKilobytes)
			d.Set("requires_duplicate_detection", props.RequiresDuplicateDetection)
			d.Set("support_ordering", props.SupportOrdering)

			if maxSizeMB := props.MaxSizeInMegabytes; maxSizeMB != nil {
				maxSize := int(*props.MaxSizeInMegabytes)

				// if the topic is in a premium namespace and partitioning is enabled then the
				// max size returned by the API will be 16 times greater than the value set
				if partitioning := props.EnablePartitioning; partitioning != nil && *partitioning {
					namespacesClient := meta.(*clients.Client).ServiceBus.NamespacesClient
					namespaceId := namespaces.NewNamespaceID(id.SubscriptionId, id.ResourceGroupName, id.NamespaceName)
					namespaceResp, err := namespacesClient.Get(ctx, namespaceId)
					if err != nil {
						return err
					}

					if namespaceModel := namespaceResp.Model; namespaceModel != nil {
						if namespaceModel.Sku.Name != namespaces.SkuNamePremium {
							const partitionCount = 16
							maxSize = int(*props.MaxSizeInMegabytes / partitionCount)
						}
					}
				}

				d.Set("max_size_in_megabytes", maxSize)
			}
		}
	}

	return nil
}

func resourceServiceBusTopicDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ServiceBus.TopicsClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := topics.ParseTopicID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Delete(ctx, *id)
	if err != nil {
		if !response.WasNotFound(resp.HttpResponse) {
			return fmt.Errorf("deleting %s: %+v", id, err)
		}
	}

	return nil
}
