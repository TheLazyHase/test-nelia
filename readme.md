Goal
====

gate is a lightweight software gate, who apply rules to filter incoming traffic and masking rules to mask confidential or personal information on outgoing traffic.

Compilation
===========

The gate binary is built with :
> go build ./cmd/gate.go

There is also a small binary for a mock software server :
> go build ./test_service/repeater.go

this binary listen to port 4500 and repeat back at you the content of POST requests.

Use
===

You need a conf.json file where the binary is, which will contain all the filtering and masking rules. The conf.json.orig is a sample one. The service will fail on startup if there is no config.json or if it's malformed

```json
{
	"gated_url": "http://localhost:4500",
	"header_rules": [
		{
			"key": "Myheader",
			"match": ".*",
			"name": "Test rule disabling Myheader"
		}
	],
	"parameter_rules": [
		{
			"key": "user",
			"match": ".*",
			"name": "Test rule disabling URL parameter user"
		}
	],
	"segment_rules": [
		{
			"key": "ODS",
			"name": "Test rule disabling messages with an ODS segment"
		}
	],
	"masking_rules": [
		{
			"key": "OBX",
			"fields": [7, 11],
			"name": "Test rule disabling some observation fields on OBX"
		},
		{
			"key": "ODT",
			"everything": true,
			"name": "Test rule disabling all observation fields on ODT"
		}
	]
}
```

gated_url is mandatory and is the adress of the service you want to gate.

header_rules are a list of rules that deny incoming traffic based on HTTP headers. Each rules need a key which is the header name that the filter will apply on, a match which is a RE2 regexp applied to the value of the header, and a name that will be displayed when a request is denied due to the rule. Any request with an header who match the rule will be denied.

parameter_rules are similarly applied to URL parameters. They need a key which is the URL parameter name, a match who is a RE2 regexp that need to match the value, and a name.

See https://github.com/google/re2/wiki/Syntax for RE2 syntax.

segment_rules will deny incoming traffic whose body is a .hl7 message with a segment that match the rule. No match rules, only the presence of the segment is considered.

masking_rules allow to mask all or part of the .hl7 results returned by the software. Each rule have a key, which is a segment name, and a list of field IDs. Alternatively, you can set "everything" at true which will match all field IDs. Matching fields are replaced by U (unknown) in the output.

If several rules are written for the same key, only one will be applied. We provide no guarantee on which one will be applied.

Note that the gate is hardcoded to filter every field of PID (patient identification) segments. This cannot be overrided.

If you change the conf.json file, the gate will automatically apply the new configuration. If the conf.json is incorrect, the gate will deny all request until correct configuration is provided.

The gate also log every requests, including content and headers, on standard log output.

Future evolutions
=================

* tests lack substances, and while it's mostly a case of copying and expanding existing one, there's still not enough unit or integration tests
* While this is following the instruction, it should be white-list based and not black-list based
* There is a lot of potential use for finer rules, like stripping bad headers instead of denying requests entirely, or rules on http method
* schema-based rules for .hl7 would make sense.
* The code use the github.com/esequiel378/hl7 library to parse hl7 content, but this library seem underwhelming and in particular don't validate content very well.
* expanding test suites would be good, especially more specific tests. Examples are from https://docs.webchartnow.com/features/system-administration/interfaces/sample-hl7-messages/ but without hand on experience it's hard to say if they are well chosen.
* code division in middleware probably should be thought about again
