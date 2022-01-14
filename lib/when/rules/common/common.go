package common

import "github.com/mrbentarikau/pagst/lib/when/rules"

var All = []rules.Rule{
	SlashDMY(rules.Override),
}
