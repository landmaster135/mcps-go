package sequentialthinking

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	mcp "github.com/mark3labs/mcp-go/mcp"
	server "github.com/mark3labs/mcp-go/server"
)

const (
	version = "0.2.0"
)

// ThoughtData は思考データを表す構造体
type ThoughtData struct {
	Thought           string `json:"thought"`
	ThoughtNumber     int    `json:"thoughtNumber"`
	TotalThoughts     int    `json:"totalThoughts"`
	NextThoughtNeeded bool   `json:"nextThoughtNeeded"`
	IsRevision        bool   `json:"isRevision,omitempty"`
	RevisesThought    int    `json:"revisesThought,omitempty"`
	BranchFromThought int    `json:"branchFromThought,omitempty"`
	BranchID          string `json:"branchId,omitempty"`
	NeedsMoreThoughts bool   `json:"needsMoreThoughts,omitempty"`
}

// SequentialThinkingServer はシーケンシャルシンキングサーバーを表す構造体
type SequentialThinkingServer struct {
	ThoughtHistory []ThoughtData
	Branches       map[string][]ThoughtData
}

// NewSequentialThinkingServer は新しいシーケンシャルシンキングサーバーを作成します
func NewSequentialThinkingServer() *SequentialThinkingServer {
	return &SequentialThinkingServer{
		ThoughtHistory: []ThoughtData{},
		Branches:       make(map[string][]ThoughtData),
	}
}

// ValidateThoughtData は思考データを検証します
func (s *SequentialThinkingServer) ValidateThoughtData(args map[string]interface{}) (ThoughtData, error) {
	thought, ok := args["thought"].(string)
	if !ok || thought == "" {
		return ThoughtData{}, fmt.Errorf("invalid thought: must be a string")
	}

	thoughtNumber, ok := args["thoughtNumber"].(float64)
	if !ok || thoughtNumber < 1 {
		return ThoughtData{}, fmt.Errorf("invalid thoughtNumber: must be a number greater than 0")
	}

	totalThoughts, ok := args["totalThoughts"].(float64)
	if !ok || totalThoughts < 1 {
		return ThoughtData{}, fmt.Errorf("invalid totalThoughts: must be a number greater than 0")
	}

	nextThoughtNeeded, ok := args["nextThoughtNeeded"].(bool)
	if !ok {
		return ThoughtData{}, fmt.Errorf("invalid nextThoughtNeeded: must be a boolean")
	}

	data := ThoughtData{
		Thought:           thought,
		ThoughtNumber:     int(thoughtNumber),
		TotalThoughts:     int(totalThoughts),
		NextThoughtNeeded: nextThoughtNeeded,
	}

	// オプションフィールドの処理
	if isRevision, ok := args["isRevision"].(bool); ok {
		data.IsRevision = isRevision
	}

	if revisesThought, ok := args["revisesThought"].(float64); ok {
		data.RevisesThought = int(revisesThought)
	}

	if branchFromThought, ok := args["branchFromThought"].(float64); ok {
		data.BranchFromThought = int(branchFromThought)
	}

	if branchID, ok := args["branchId"].(string); ok {
		data.BranchID = branchID
	}

	if needsMoreThoughts, ok := args["needsMoreThoughts"].(bool); ok {
		data.NeedsMoreThoughts = needsMoreThoughts
	}

	return data, nil
}

// FormatThought は思考データをフォーマットします
func (s *SequentialThinkingServer) FormatThought(data ThoughtData) string {
	var prefix, context string

	if data.IsRevision {
		prefix = "🔄 Revision"
		context = fmt.Sprintf(" (revising thought %d)", data.RevisesThought)
	} else if data.BranchFromThought > 0 {
		prefix = "🌿 Branch"
		context = fmt.Sprintf(" (from thought %d, ID: %s)", data.BranchFromThought, data.BranchID)
	} else {
		prefix = "💭 Thought"
		context = ""
	}

	header := fmt.Sprintf("%s %d/%d%s", prefix, data.ThoughtNumber, data.TotalThoughts, context)
	border := "─"
	borderLength := max(len(header), len(data.Thought)) + 4
	borderLine := "┌" + repeatString(border, borderLength) + "┐"

	return fmt.Sprintf(`
%s
│ %s │
├%s┤
│ %s │
└%s┘`,
		borderLine,
		header,
		repeatString(border, borderLength),
		data.Thought,
		repeatString(border, borderLength))
}

// ProcessThought は思考データを処理します
func (s *SequentialThinkingServer) ProcessThought(args map[string]interface{}) (*mcp.CallToolResult, error) {
	data, err := s.ValidateThoughtData(args)
	if err != nil {
		return nil, err
	}

	// 思考番号が総思考数を超える場合、総思考数を更新
	if data.ThoughtNumber > data.TotalThoughts {
		data.TotalThoughts = data.ThoughtNumber
	}

	// 思考履歴に追加
	s.ThoughtHistory = append(s.ThoughtHistory, data)

	// ブランチ処理
	if data.BranchFromThought > 0 && data.BranchID != "" {
		if _, ok := s.Branches[data.BranchID]; !ok {
			s.Branches[data.BranchID] = []ThoughtData{}
		}
		s.Branches[data.BranchID] = append(s.Branches[data.BranchID], data)
	}

	// フォーマットされた思考を表示
	formattedThought := s.FormatThought(data)
	fmt.Fprintln(os.Stderr, formattedThought)

	// 結果を返す
	result := map[string]interface{}{
		"thoughtNumber":      data.ThoughtNumber,
		"totalThoughts":      data.TotalThoughts,
		"nextThoughtNeeded":  data.NextThoughtNeeded,
		"branches":           getKeys(s.Branches),
		"thoughtHistoryLength": len(s.ThoughtHistory),
	}

	jsonResult, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, err
	}

	return mcp.NewToolResultText(string(jsonResult)), nil
}

// HandleSequentialThinking はシーケンシャルシンキングツールのハンドラー
func (s *SequentialThinkingServer) HandleSequentialThinking(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.ProcessThought(request.Params.Arguments)
}

// BuildSequentialThinkingServer はシーケンシャルシンキングのMCPサーバーを構築します
func BuildSequentialThinkingServer() {
	s := createSequentialThinkingServer()

	// サーバーを起動
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

func createSequentialThinkingServer() *server.MCPServer {
	// MCPサーバーを作成
	s := server.NewMCPServer(
		"Sequential Thinking Server",
		version,
		server.WithResourceCapabilities(false, false),
		server.WithLogging(),
	)

	// シーケンシャルシンキングサーバーを作成
	thinkingServer := NewSequentialThinkingServer()

	// ツール: シーケンシャルシンキング
	sequentialThinkingTool := mcp.NewTool("sequentialthinking",
		mcp.WithDescription(`A detailed tool for dynamic and reflective problem-solving through thoughts.
This tool helps analyze problems through a flexible thinking process that can adapt and evolve.
Each thought can build on, question, or revise previous insights as understanding deepens.

When to use this tool:
- Breaking down complex problems into steps
- Planning and design with room for revision
- Analysis that might need course correction
- Problems where the full scope might not be clear initially
- Problems that require a multi-step solution
- Tasks that need to maintain context over multiple steps
- Situations where irrelevant information needs to be filtered out

Key features:
- You can adjust total_thoughts up or down as you progress
- You can question or revise previous thoughts
- You can add more thoughts even after reaching what seemed like the end
- You can express uncertainty and explore alternative approaches
- Not every thought needs to build linearly - you can branch or backtrack
- Generates a solution hypothesis
- Verifies the hypothesis based on the Chain of Thought steps
- Repeats the process until satisfied
- Provides a correct answer

Parameters explained:
- thought: Your current thinking step, which can include:
* Regular analytical steps
* Revisions of previous thoughts
* Questions about previous decisions
* Realizations about needing more analysis
* Changes in approach
* Hypothesis generation
* Hypothesis verification
- next_thought_needed: True if you need more thinking, even if at what seemed like the end
- thought_number: Current number in sequence (can go beyond initial total if needed)
- total_thoughts: Current estimate of thoughts needed (can be adjusted up/down)
- is_revision: A boolean indicating if this thought revises previous thinking
- revises_thought: If is_revision is true, which thought number is being reconsidered
- branch_from_thought: If branching, which thought number is the branching point
- branch_id: Identifier for the current branch (if any)
- needs_more_thoughts: If reaching end but realizing more thoughts needed

You should:
1. Start with an initial estimate of needed thoughts, but be ready to adjust
2. Feel free to question or revise previous thoughts
3. Don't hesitate to add more thoughts if needed, even at the "end"
4. Express uncertainty when present
5. Mark thoughts that revise previous thinking or branch into new paths
6. Ignore information that is irrelevant to the current step
7. Generate a solution hypothesis when appropriate
8. Verify the hypothesis based on the Chain of Thought steps
9. Repeat the process until satisfied with the solution
10. Provide a single, ideally correct answer as the final output
11. Only set next_thought_needed to false when truly done and a satisfactory answer is reached`),
		mcp.WithString("thought",
			mcp.Required(),
			mcp.Description("Your current thinking step"),
		),
		mcp.WithBoolean("nextThoughtNeeded",
			mcp.Required(),
			mcp.Description("Whether another thought step is needed"),
		),
		mcp.WithNumber("thoughtNumber",
			mcp.Required(),
			mcp.Description("Current thought number"),
			mcp.Min(1),
		),
		mcp.WithNumber("totalThoughts",
			mcp.Required(),
			mcp.Description("Estimated total thoughts needed"),
			mcp.Min(1),
		),
		mcp.WithBoolean("isRevision",
			mcp.Description("Whether this revises previous thinking"),
		),
		mcp.WithNumber("revisesThought",
			mcp.Description("Which thought is being reconsidered"),
			mcp.Min(1),
		),
		mcp.WithNumber("branchFromThought",
			mcp.Description("Branching point thought number"),
			mcp.Min(1),
		),
		mcp.WithString("branchId",
			mcp.Description("Branch identifier"),
		),
		mcp.WithBoolean("needsMoreThoughts",
			mcp.Description("If more thoughts are needed"),
		),
	)

	s.AddTool(sequentialThinkingTool, thinkingServer.HandleSequentialThinking)

	return s
}

// ヘルパー関数

// max は2つの整数の最大値を返します
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// repeatString は文字列を指定された回数繰り返します
func repeatString(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}

// getKeys はマップのキーをスライスとして返します
func getKeys(m map[string][]ThoughtData) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
