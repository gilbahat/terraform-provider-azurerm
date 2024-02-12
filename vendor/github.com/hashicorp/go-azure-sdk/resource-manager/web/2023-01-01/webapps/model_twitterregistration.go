package webapps

// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See NOTICE.txt in the project root for license information.

type TwitterRegistration struct {
	ConsumerKey               *string `json:"consumerKey,omitempty"`
	ConsumerSecretSettingName *string `json:"consumerSecretSettingName,omitempty"`
}
