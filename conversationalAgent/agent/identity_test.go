package agent

import "testing"

func TestIsIdentityQuestion(t *testing.T) {
	// These should all be detected as identity questions
	positives := []string{
		// Exact matches
		"What can you do?",
		"what can you do",
		"Who are you?",
		"who are you",
		"Help",
		"help me",
		"What are your capabilities?",
		"How can you help me?",
		"Tell me about yourself",
		"What tools do you have?",
		"What do you support?",
		"What can I ask you?",
		"Introduce yourself",
		"What all can you do?",

		// Rephrasings (keyword-group matching)
		"What are your features and capabilities?",
		"Tell me about your capabilities",
		"What is this assistant capable of?",
		"What can this bot do for me?",
		"Describe your abilities",
		"What is the purpose of this agent?",
		"What functions do you offer?",
		"How can you assist with my problems?",
		"What kind of support do you provide?",
		"Are you able to help with alerts?",
		"What powers does this autopilot have?",
		"List your skills",

		// Punctuation variants
		"What can you do!",
		"What can you do...",
		"  what can you do?  ",
	}

	for _, q := range positives {
		if !isIdentityQuestion(q) {
			t.Errorf("Expected TRUE for identity question: %q", q)
		}
	}

	// These should NOT be detected as identity questions
	negatives := []string{
		"Show me all critical alerts",
		"List all AWS resources",
		"Give me an environment summary",
		"Why is the GreenLake portal slow?",
		"Investigate web-server-prod-01",
		"What is the runbook for high CPU?",
		"Predict capacity for db-primary-01",
		"Show open incidents",
		"Correlate network for greenlake-portal",
		"What is the blast radius for k8s-node-04?",
		"How many resources do we have?",
		"Are there any urgent incidents?",
		"Hello",
		"Good morning",
		"Thanks",
	}

	for _, q := range negatives {
		if isIdentityQuestion(q) {
			t.Errorf("Expected FALSE for infrastructure question: %q", q)
		}
	}
}
