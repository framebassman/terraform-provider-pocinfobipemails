// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// Ensure interface compliance.
var _ planmodifier.String = htmlWhitespaceInsensitiveModifier{}

// htmlWhitespaceInsensitiveModifier suppresses diffs when only whitespace differs.
type htmlWhitespaceInsensitiveModifier struct{}

func (m htmlWhitespaceInsensitiveModifier) Description(ctx context.Context) string {
	return "Ignores insignificant whitespace differences in HTML content."
}

func (m htmlWhitespaceInsensitiveModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m htmlWhitespaceInsensitiveModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() || req.PlanValue.IsNull() {
		return
	}

	normalize := func(raw string) string {
		s := strings.ReplaceAll(raw, "\r\n", "\n")
		s = strings.TrimSpace(s)
		s = strings.Join(strings.Fields(s), " ")
		re := regexp.MustCompile(`>[\s]*<`)
		s = re.ReplaceAllString(s, "><")
		return s
	}

	oldVal := normalize(req.StateValue.ValueString())
	newVal := normalize(req.PlanValue.ValueString())

	if oldVal == newVal {
		resp.PlanValue = req.StateValue
	}
}
