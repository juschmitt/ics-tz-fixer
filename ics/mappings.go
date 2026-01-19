package ics

type tzSignature struct {
	iana string
	rule tzRule
}

var knownSignatures = []tzSignature{
	{
		iana: "Europe/Berlin",
		rule: tzRule{
			standardOffsetFrom: "+0200",
			standardOffsetTo:   "+0100",
			standardRRule:      normalizeRRule("FREQ=YEARLY;INTERVAL=1;BYDAY=-1SU;BYMONTH=10"),
			daylightOffsetFrom: "+0100",
			daylightOffsetTo:   "+0200",
			daylightRRule:      normalizeRRule("FREQ=YEARLY;INTERVAL=1;BYDAY=-1SU;BYMONTH=3"),
		},
	},
	{
		iana: "Europe/London",
		rule: tzRule{
			standardOffsetFrom: "+0100",
			standardOffsetTo:   "+0000",
			standardRRule:      normalizeRRule("FREQ=YEARLY;INTERVAL=1;BYDAY=-1SU;BYMONTH=10"),
			daylightOffsetFrom: "+0000",
			daylightOffsetTo:   "+0100",
			daylightRRule:      normalizeRRule("FREQ=YEARLY;INTERVAL=1;BYDAY=-1SU;BYMONTH=3"),
		},
	},
	{
		iana: "America/New_York",
		rule: tzRule{
			standardOffsetFrom: "-0400",
			standardOffsetTo:   "-0500",
			standardRRule:      normalizeRRule("FREQ=YEARLY;BYMONTH=11;BYDAY=1SU"),
			daylightOffsetFrom: "-0500",
			daylightOffsetTo:   "-0400",
			daylightRRule:      normalizeRRule("FREQ=YEARLY;BYMONTH=3;BYDAY=2SU"),
		},
	},
}

var knownTZIDOverrides = map[string]string{
	"Romance Standard Time": "Europe/Berlin",
}

func knownSignatureMapping() map[string]string {
	mapping := make(map[string]string, len(knownSignatures))
	for _, signature := range knownSignatures {
		mapping[signature.rule.signature()] = signature.iana
	}
	return mapping
}

func knownTZIDMapping() map[string]string {
	mapping := make(map[string]string, len(knownTZIDOverrides))
	for tzid, iana := range knownTZIDOverrides {
		mapping[tzid] = iana
	}
	return mapping
}
