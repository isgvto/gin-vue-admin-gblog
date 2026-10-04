package blog

import (
	"strings"
	"testing"
)

func TestCompactFlowchartLimits(t *testing.T) {
	tests := []struct {
		name, source string
		valid        bool
	}{
		{"compact feedback", `flowchart TD
 A["写作"] --> B{"达标？"}
 B -->|"否"| A
 B -->|"是"| C["发布"]`, true},
		{"inline classes", `flowchart TD
 A["写作"]:::step --> B["保存"]:::step
 classDef step fill:#eff6ff,stroke:#60a5fa;`, true},
		{"long chain", `flowchart TD
 A --> B --> C --> D --> E --> F --> G`, true},
		{"numeric IDs", "flowchart TD\n1 --> 2 --> 3 --> 4 --> 5 --> 6 --> 7 --> 8 --> 9", true},
		{"labeled long chain", `flowchart TD
 A -- "yes" --> B -- "yes" --> C -- "yes" --> D -- "yes" --> E -- "yes" --> F -- "yes" --> G`, true},
		{"nine nodes retained", `flowchart TD
 A --> B
 A --> C
 A --> D
 A --> E
 A --> F
 A --> G
 A --> H
 A --> I`, true},
		{"group frame", "flowchart TD\nsubgraph phase[创作]\nA --> B\nend", true},
		{"long label", `flowchart TD
 A["这是一个包含过多说明文字不适合作为图示节点的标签详细说明详细说明详细说明详细说明详细说明"] --> B["完成"]`, false},
		{"nested groups", "flowchart TD\nsubgraph one\nsubgraph two\nA --> B\nend\nend", false},
		{"unclosed group", "flowchart TD\nsubgraph one\nA --> B", false},
		{"too many groups", "flowchart TD\n" + strings.Repeat("subgraph stage\nA --> B\nend\n", 5), false},
		{"too many nodes", "flowchart TD\n" + strings.ReplaceAll("A B C D E F G H I J K L M N O P Q R S T U V W X Y", " ", " --> "), false},
		{"wrapped label", "flowchart TD\nA[\"`准备正文材料\n核对事实来源`\"] --> B[\"撰写\"]", true},
		{"too many edges", "flowchart TD\n" + strings.Repeat("A --> B\n", 41), false},
	}
	for _, item := range tests {
		t.Run(item.name, func(t *testing.T) {
			err := ValidateVisualPlan(VisualPlan{Kind: "flowchart", Prompt: "正文流程", Mermaid: item.source}, "flowchart")
			if (err == nil) != item.valid {
				t.Fatalf("valid=%v, error=%v", item.valid, err)
			}
		})
	}
	if err := ValidateVisualPlan(VisualPlan{Kind: "structure", Prompt: "模块关系", Mermaid: "flowchart TD\nsubgraph phase[模块]\nA --> B\nend"}, "structure"); err != nil {
		t.Fatalf("structure flow changed: %v", err)
	}
}
