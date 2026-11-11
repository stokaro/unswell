package builtin

// This file holds phrase lists qualified on the assistant-panel corpus.
// Two assistants from different model families reviewed each finding under
// the statement of its rule. A phrase is listed when both accepted at least
// 80% of at least five reviewed findings of its word family.
// docs/research/rule-qualification.md records the measurement.

// qualifiedPraise lists general praise that names no property or measurement.
func qualifiedPraise() []string {
	return []string{
		"beautiful", "beautifully", "comprehensive documentation", "comprehensive guide", "easy to use",
		"excellent", "exceptional", "exceptionally", "intuitive", "powerful", "seamless",
		"seamless integration", "seamlessly", "sophisticated", "state-of-the-art", "unprecedented",
	}
}

// qualifiedClusterWords lists evaluative modifiers that count toward a cluster.
// One alone is not reported by the cluster rule.
func qualifiedClusterWords() []string {
	return []string{
		"beautiful", "comprehensive", "crucial", "cutting-edge", "delightful", "easy-to-use",
		"enterprise-grade", "excellent", "extensive", "feature-rich", "intuitive", "powerful", "robust",
		"seamless", "seamlessly", "significantly", "sophisticated", "state-of-the-art",
		"straightforward", "streamlined", "unprecedented", "user-friendly",
	}
}

// qualifiedWordyPhrases lists stock verbs and phrases that a plain word replaces.
func qualifiedWordyPhrases() []string {
	return []string{
		"by adhering to", "empower", "empowering", "empowers", "facilitate", "facilitates",
		"facilitating", "first and foremost", "foster", "fostering", "fosters", "leverage", "leveraged",
		"leverages", "leveraging", "the bottom line", "utilize", "utilizes", "utilizing",
	}
}
