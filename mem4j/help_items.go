package main

type helpItem struct {
	name    string
	summary string
}

var resourceHelpItems = []helpItem{
	{"scope", "Manage scope hubs and scope-to-scope links"},
	{"instruction", "Manage always-on instructions and recall merged output"},
	{"reminder", "Manage prompt-triggered instant reminders with embedding recall"},
	{"entity", "Manage entity hubs, entity-to-entity links, and neighbor traversal"},
	{"statement", "Manage current facts linked to entities with embedding recall"},
	{"event", "Manage timestamped history linked to entities with embedding recall"},
	{"graph", "Run read-only Cypher queries against the Neo4j graph"},
	{"hook", "Handle Claude Code and agy CLI hook events"},
}

var actionHelpItems = map[string][]helpItem{
	"scope": {
		{"list", "List all scopes with descriptions"},
		{"create", "Create a new scope"},
		{"edit", "Update a scope description"},
		{"link", "Create a relationship between two scopes"},
		{"unlink", "Delete a relationship between two scopes"},
		{"neighbors", "Traverse connected scopes up to depth 1..3"},
		{"delete", "Delete a scope with no linked instructions or reminders"},
	},
	"instruction": {
		{"recall", "Merge instructions matching the intersection of active scopes"},
		{"create", "Create an instruction linked to one or more scopes"},
		{"edit", "Update an instruction title or body"},
		{"link", "Link an instruction to an additional scope"},
		{"unlink", "Unlink an instruction from a scope"},
		{"delete", "Delete an instruction"},
	},
	"reminder": {
		{"recall", "Find reminders matching active scopes and prompt embedding similarity"},
		{"create", "Create a reminder and embed its trigger description"},
		{"edit", "Update a reminder title, description (re-embedded), or body"},
		{"link", "Link a reminder to an additional scope"},
		{"unlink", "Unlink a reminder from a scope"},
		{"delete", "Delete a reminder"},
	},
	"entity": {
		{"list", "List all entities with descriptions"},
		{"recall", "Find entities by embedding similarity"},
		{"create", "Create an entity (or promote matching scope) with embedding"},
		{"edit", "Update an entity description and re-embed"},
		{"link", "Create a relationship between two entities"},
		{"unlink", "Delete a relationship between two entities"},
		{"neighbors", "Traverse connected entities and 1-hop statements/events"},
		{"delete", "Delete an entity with no linked statements or events"},
	},
	"statement": {
		{"list", "List 1-hop statements linked to an entity"},
		{"recall", "Find statements by embedding similarity (optionally filtered by entity)"},
		{"create", "Create a statement linked to one or more entities with embedding"},
		{"edit", "Update a statement title, description (re-embedded), or body"},
		{"link", "Link a statement to an additional entity"},
		{"unlink", "Unlink a statement from an entity"},
		{"delete", "Delete a statement"},
	},
	"event": {
		{"list", "List 1-hop events linked to an entity (newest first)"},
		{"recall", "Find events by embedding similarity (optionally filtered by entity)"},
		{"create", "Create a timestamped event linked to one or more entities with embedding"},
		{"edit", "Update an event title, description (re-embedded), timestamp, or body"},
		{"link", "Link an event to an additional entity"},
		{"unlink", "Unlink an event from an entity"},
		{"delete", "Delete an event"},
	},
	"graph": {
		{"query", "Execute a read-only Cypher query and print JSON Lines"},
	},
	"hook": {
		{"claude-session-start", "Inject active instructions and scope list on Claude Code SessionStart"},
		{"claude-user-prompt", "Inject matching reminders, statements, and events on Claude Code UserPromptSubmit"},
		{"agy-session-start", "Inject active instructions and scope list on agy SessionStart"},
		{"agy-pre-invocation", "Inject matching reminders, statements, and events on agy PreInvocation (once per user turn)"},
	},
}
