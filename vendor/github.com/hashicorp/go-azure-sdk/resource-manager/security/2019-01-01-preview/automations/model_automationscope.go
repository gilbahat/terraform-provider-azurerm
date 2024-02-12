package automations

// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See NOTICE.txt in the project root for license information.

type AutomationScope struct {
	Description *string `json:"description,omitempty"`
	ScopePath   *string `json:"scopePath,omitempty"`
}
