package figma

import (
	"fmt"
	"time"
)

// SimplifiedDesign はFigmaデザインの簡略化された構造を表します
type SimplifiedDesign struct {
	Name         string           `json:"name"`
	LastModified string           `json:"lastModified"`
	ThumbnailURL string           `json:"thumbnailUrl"`
	Nodes        []SimplifiedNode `json:"nodes"`
	GlobalVars   GlobalVars       `json:"globalVars"`
}

// SimplifiedNode はFigmaノードの簡略化された構造を表します
type SimplifiedNode struct {
	ID           string           `json:"id"`
	Name         string           `json:"name"`
	Type         string           `json:"type"`
	BoundingBox  *BoundingBox     `json:"boundingBox,omitempty"`
	Text         string           `json:"text,omitempty"`
	TextStyle    string           `json:"textStyle,omitempty"`
	Fills        string           `json:"fills,omitempty"`
	Styles       string           `json:"styles,omitempty"`
	Strokes      string           `json:"strokes,omitempty"`
	Effects      string           `json:"effects,omitempty"`
	Opacity      float64          `json:"opacity,omitempty"`
	BorderRadius string           `json:"borderRadius,omitempty"`
	Layout       string           `json:"layout,omitempty"`
	Children     []SimplifiedNode `json:"children,omitempty"`
}

// BoundingBox はノードの位置とサイズを表します
type BoundingBox struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// TextStyle はテキストスタイルを表します
type TextStyle struct {
	FontFamily        string  `json:"fontFamily,omitempty"`
	FontWeight        float64 `json:"fontWeight,omitempty"`
	FontSize          float64 `json:"fontSize,omitempty"`
	LineHeight        string  `json:"lineHeight,omitempty"`
	LetterSpacing     string  `json:"letterSpacing,omitempty"`
	TextCase          string  `json:"textCase,omitempty"`
	TextAlignHorizontal string  `json:"textAlignHorizontal,omitempty"`
	TextAlignVertical   string  `json:"textAlignVertical,omitempty"`
}

// SimplifiedFill は塗りつぶしの簡略化された構造を表します
type SimplifiedFill struct {
	Type                  string    `json:"type,omitempty"`
	Hex                   string    `json:"hex,omitempty"`
	RGBA                  string    `json:"rgba,omitempty"`
	Opacity               float64   `json:"opacity,omitempty"`
	ImageRef              string    `json:"imageRef,omitempty"`
	ScaleMode             string    `json:"scaleMode,omitempty"`
	GradientHandlePositions []Vector  `json:"gradientHandlePositions,omitempty"`
	GradientStops         []GradientStop `json:"gradientStops,omitempty"`
}

// Vector はベクトル座標を表します
type Vector struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// GradientStop はグラデーションの停止点を表します
type GradientStop struct {
	Position float64    `json:"position"`
	Color    ColorValue `json:"color"`
}

// ColorValue は色の値を表します
type ColorValue struct {
	Hex     string  `json:"hex"`
	Opacity float64 `json:"opacity"`
}

// SimplifiedStroke はストロークの簡略化された構造を表します
type SimplifiedStroke struct {
	Colors []string `json:"colors"`
	Weight float64  `json:"weight"`
	Align  string   `json:"align,omitempty"`
}

// SimplifiedEffects はエフェクトの簡略化された構造を表します
type SimplifiedEffects struct {
	Shadows []Shadow `json:"shadows,omitempty"`
	Blur    *Blur    `json:"blur,omitempty"`
}

// Shadow は影の効果を表します
type Shadow struct {
	Type     string  `json:"type"`
	Color    string  `json:"color"`
	OffsetX  float64 `json:"offsetX"`
	OffsetY  float64 `json:"offsetY"`
	Radius   float64 `json:"radius"`
	Spread   float64 `json:"spread,omitempty"`
	Visible  bool    `json:"visible"`
	BlendMode string  `json:"blendMode,omitempty"`
}

// Blur はぼかし効果を表します
type Blur struct {
	Type   string  `json:"type"`
	Radius float64 `json:"radius"`
}

// SimplifiedLayout はレイアウトの簡略化された構造を表します
type SimplifiedLayout struct {
	Position      string  `json:"position,omitempty"`
	Width         float64 `json:"width,omitempty"`
	Height        float64 `json:"height,omitempty"`
	PaddingLeft   float64 `json:"paddingLeft,omitempty"`
	PaddingRight  float64 `json:"paddingRight,omitempty"`
	PaddingTop    float64 `json:"paddingTop,omitempty"`
	PaddingBottom float64 `json:"paddingBottom,omitempty"`
	SpacingH      float64 `json:"spacingH,omitempty"`
	SpacingV      float64 `json:"spacingV,omitempty"`
	Direction     string  `json:"direction,omitempty"`
	LayoutMode    string  `json:"layoutMode,omitempty"`
	LayoutAlign   string  `json:"layoutAlign,omitempty"`
	LayoutGrow    float64 `json:"layoutGrow,omitempty"`
}

// GlobalVars はグローバル変数を表します
type GlobalVars struct {
	Styles map[string]interface{} `json:"styles"`
}

// FigmaError はFigma APIのエラーを表します
type FigmaError struct {
	Status int    `json:"status"`
	Err    string `json:"err"`
}

// FigmaFileResponse はFigma APIのファイルレスポンスを表します
type FigmaFileResponse struct {
	Name         string                 `json:"name"`
	LastModified string                 `json:"lastModified"`
	ThumbnailURL string                 `json:"thumbnailUrl"`
	Version      string                 `json:"version"`
	Document     FigmaDocumentNode      `json:"document"`
	Components   map[string]interface{} `json:"components"`
	Styles       map[string]interface{} `json:"styles"`
}

// FigmaFileNodesResponse はFigma APIのノードレスポンスを表します
type FigmaFileNodesResponse struct {
	Name         string                        `json:"name"`
	LastModified string                        `json:"lastModified"`
	ThumbnailURL string                        `json:"thumbnailUrl,omitempty"`
	Nodes        map[string]FigmaNodeResponse  `json:"nodes"`
}

// FigmaNodeResponse はFigma APIのノードレスポンスを表します
type FigmaNodeResponse struct {
	Document FigmaDocumentNode `json:"document"`
}

// FigmaDocumentNode はFigmaドキュメントノードを表します
type FigmaDocumentNode struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	Type        string             `json:"type"`
	Visible     *bool              `json:"visible,omitempty"`
	Children    []FigmaDocumentNode `json:"children,omitempty"`
	Fills       []FigmaFill        `json:"fills,omitempty"`
	Strokes     []FigmaFill        `json:"strokes,omitempty"`
	StrokeWeight float64            `json:"strokeWeight,omitempty"`
	StrokeAlign  string             `json:"strokeAlign,omitempty"`
	Effects     []FigmaEffect      `json:"effects,omitempty"`
	Style       *FigmaStyle        `json:"style,omitempty"`
	Characters  string             `json:"characters,omitempty"`
	Opacity     float64            `json:"opacity,omitempty"`
	CornerRadius float64           `json:"cornerRadius,omitempty"`
	RectangleCornerRadii []float64 `json:"rectangleCornerRadii,omitempty"`
	LayoutMode  string             `json:"layoutMode,omitempty"`
	PrimaryAxisSizingMode string   `json:"primaryAxisSizingMode,omitempty"`
	CounterAxisSizingMode string   `json:"counterAxisSizingMode,omitempty"`
	PrimaryAxisAlignItems string   `json:"primaryAxisAlignItems,omitempty"`
	CounterAxisAlignItems string   `json:"counterAxisAlignItems,omitempty"`
	PaddingLeft  float64           `json:"paddingLeft,omitempty"`
	PaddingRight float64           `json:"paddingRight,omitempty"`
	PaddingTop   float64           `json:"paddingTop,omitempty"`
	PaddingBottom float64          `json:"paddingBottom,omitempty"`
	ItemSpacing  float64           `json:"itemSpacing,omitempty"`
	LayoutAlign  string            `json:"layoutAlign,omitempty"`
	LayoutGrow   float64           `json:"layoutGrow,omitempty"`
	AbsoluteBoundingBox *BoundingBox `json:"absoluteBoundingBox,omitempty"`
}

// FigmaFill はFigmaの塗りつぶしを表します
type FigmaFill struct {
	Type                  string    `json:"type"`
	Visible               bool      `json:"visible"`
	Opacity               float64   `json:"opacity,omitempty"`
	Color                 *FigmaColor `json:"color,omitempty"`
	BlendMode             string    `json:"blendMode,omitempty"`
	GradientHandlePositions []Vector  `json:"gradientHandlePositions,omitempty"`
	GradientStops         []FigmaGradientStop `json:"gradientStops,omitempty"`
	ScaleMode             string    `json:"scaleMode,omitempty"`
	ImageRef              string    `json:"imageRef,omitempty"`
}

// FigmaColor はFigmaの色を表します
type FigmaColor struct {
	R float64 `json:"r"`
	G float64 `json:"g"`
	B float64 `json:"b"`
	A float64 `json:"a,omitempty"`
}

// FigmaGradientStop はFigmaのグラデーション停止点を表します
type FigmaGradientStop struct {
	Position float64    `json:"position"`
	Color    FigmaColor `json:"color"`
}

// FigmaEffect はFigmaのエフェクトを表します
type FigmaEffect struct {
	Type     string     `json:"type"`
	Visible  bool       `json:"visible"`
	Radius   float64    `json:"radius,omitempty"`
	Color    *FigmaColor `json:"color,omitempty"`
	Offset   *Vector    `json:"offset,omitempty"`
	Spread   float64    `json:"spread,omitempty"`
	BlendMode string     `json:"blendMode,omitempty"`
}

// FigmaStyle はFigmaのスタイルを表します
type FigmaStyle struct {
	FontFamily        string  `json:"fontFamily,omitempty"`
	FontWeight        float64 `json:"fontWeight,omitempty"`
	FontSize          float64 `json:"fontSize,omitempty"`
	LineHeightPx      float64 `json:"lineHeightPx,omitempty"`
	LetterSpacing     float64 `json:"letterSpacing,omitempty"`
	TextCase          string  `json:"textCase,omitempty"`
	TextDecoration    string  `json:"textDecoration,omitempty"`
	TextAlignHorizontal string  `json:"textAlignHorizontal,omitempty"`
	TextAlignVertical   string  `json:"textAlignVertical,omitempty"`
}

// FigmaImagesResponse はFigma APIの画像レスポンスを表します
type FigmaImagesResponse struct {
	Err    string            `json:"err,omitempty"`
	Images map[string]string `json:"images"`
}

// FigmaImageFillsResponse はFigma APIの画像フィルレスポンスを表します
type FigmaImageFillsResponse struct {
	Err  string `json:"err,omitempty"`
	Meta struct {
		Images map[string]string `json:"images"`
	} `json:"meta"`
}

// FetchImageParams は画像取得パラメータを表します
type FetchImageParams struct {
	NodeID   string `json:"nodeId"`
	FileName string `json:"fileName"`
	FileType string `json:"fileType"`
}

// FetchImageFillParams は画像フィル取得パラメータを表します
type FetchImageFillParams struct {
	NodeID   string `json:"nodeId"`
	ImageRef string `json:"imageRef"`
	FileName string `json:"fileName"`
}

// ParseFigmaResponse はFigma APIのレスポンスを簡略化された構造に変換します
func ParseFigmaResponse(data interface{}) (SimplifiedDesign, error) {
	var result SimplifiedDesign
	var nodes []FigmaDocumentNode

	switch resp := data.(type) {
	case FigmaFileResponse:
		result.Name = resp.Name
		result.LastModified = resp.LastModified
		result.ThumbnailURL = resp.ThumbnailURL
		nodes = resp.Document.Children
	case FigmaFileNodesResponse:
		result.Name = resp.Name
		result.LastModified = resp.LastModified
		result.ThumbnailURL = resp.ThumbnailURL
		for _, node := range resp.Nodes {
			nodes = append(nodes, node.Document)
		}
	default:
		return result, fmt.Errorf("不明なレスポンス型です")
	}

	// グローバル変数の初期化
	result.GlobalVars = GlobalVars{
		Styles: make(map[string]interface{}),
	}

	// ノードの処理
	result.Nodes = make([]SimplifiedNode, 0, len(nodes))
	for _, node := range nodes {
		if isVisible(node) {
			simplified := parseNode(result.GlobalVars, node, FigmaDocumentNode{})
			if simplified.ID != "" {
				result.Nodes = append(result.Nodes, simplified)
			}
		}
	}

	return result, nil
}

// parseNode はFigmaノードを簡略化された構造に変換します
func parseNode(globalVars GlobalVars, node FigmaDocumentNode, parent FigmaDocumentNode) SimplifiedNode {
	simplified := SimplifiedNode{
		ID:   node.ID,
		Name: node.Name,
		Type: node.Type,
	}

	// 境界ボックスの処理
	if node.AbsoluteBoundingBox != nil {
		simplified.BoundingBox = node.AbsoluteBoundingBox
	}

	// テキストの処理
	if node.Characters != "" {
		simplified.Text = node.Characters
	}

	// スタイルの処理
	if node.Style != nil {
		textStyle := TextStyle{
			FontFamily:        node.Style.FontFamily,
			FontWeight:        node.Style.FontWeight,
			FontSize:          node.Style.FontSize,
			TextAlignHorizontal: node.Style.TextAlignHorizontal,
			TextAlignVertical:   node.Style.TextAlignVertical,
		}

		// 行の高さの計算
		if node.Style.LineHeightPx > 0 && node.Style.FontSize > 0 {
			textStyle.LineHeight = fmt.Sprintf("%.2fem", node.Style.LineHeightPx/node.Style.FontSize)
		}

		// 文字間隔の計算
		if node.Style.LetterSpacing != 0 && node.Style.FontSize > 0 {
			textStyle.LetterSpacing = fmt.Sprintf("%.2f%%", (node.Style.LetterSpacing/node.Style.FontSize)*100)
		}

		// テキストケースの設定
		textStyle.TextCase = node.Style.TextCase

		// スタイルIDの生成と設定
		styleID := generateVarID("style")
		globalVars.Styles[styleID] = textStyle
		simplified.TextStyle = styleID
	}

	// 塗りつぶしの処理
	if len(node.Fills) > 0 {
		fills := make([]SimplifiedFill, 0, len(node.Fills))
		for _, fill := range node.Fills {
			if fill.Visible {
				simplifiedFill := parseFill(fill)
				fills = append(fills, simplifiedFill)
			}
		}

		if len(fills) > 0 {
			fillID := generateVarID("fill")
			globalVars.Styles[fillID] = fills
			simplified.Fills = fillID
		}
	}

	// ストロークの処理
	if len(node.Strokes) > 0 {
		stroke := SimplifiedStroke{
			Colors: make([]string, 0, len(node.Strokes)),
			Weight: node.StrokeWeight,
			Align:  node.StrokeAlign,
		}

		for _, s := range node.Strokes {
			if s.Visible && s.Color != nil {
				color := fmt.Sprintf("rgba(%d, %d, %d, %.2f)",
					int(s.Color.R*255), int(s.Color.G*255), int(s.Color.B*255), s.Color.A)
				stroke.Colors = append(stroke.Colors, color)
			}
		}

		if len(stroke.Colors) > 0 {
			strokeID := generateVarID("stroke")
			globalVars.Styles[strokeID] = stroke
			simplified.Strokes = strokeID
		}
	}

	// エフェクトの処理
	if len(node.Effects) > 0 {
		effects := SimplifiedEffects{
			Shadows: make([]Shadow, 0),
		}

		for _, effect := range node.Effects {
			if effect.Visible {
				switch effect.Type {
				case "DROP_SHADOW", "INNER_SHADOW":
					if effect.Color != nil && effect.Offset != nil {
						shadow := Shadow{
							Type:     effect.Type,
							Color:    fmt.Sprintf("rgba(%d, %d, %d, %.2f)",
								int(effect.Color.R*255), int(effect.Color.G*255), int(effect.Color.B*255), effect.Color.A),
							OffsetX:  effect.Offset.X,
							OffsetY:  effect.Offset.Y,
							Radius:   effect.Radius,
							Spread:   effect.Spread,
							Visible:  effect.Visible,
							BlendMode: effect.BlendMode,
						}
						effects.Shadows = append(effects.Shadows, shadow)
					}
				case "LAYER_BLUR", "BACKGROUND_BLUR":
					effects.Blur = &Blur{
						Type:   effect.Type,
						Radius: effect.Radius,
					}
				}
			}
		}

		if len(effects.Shadows) > 0 || effects.Blur != nil {
			effectID := generateVarID("effect")
			globalVars.Styles[effectID] = effects
			simplified.Effects = effectID
		}
	}

	// レイアウトの処理
	layout := SimplifiedLayout{}

	// 基本的なレイアウト情報
	if node.LayoutMode != "" {
		layout.LayoutMode = node.LayoutMode
		if node.LayoutMode == "HORIZONTAL" {
			layout.Direction = "row"
		} else {
			layout.Direction = "column"
		}
		layout.PaddingLeft = node.PaddingLeft
		layout.PaddingRight = node.PaddingRight
		layout.PaddingTop = node.PaddingTop
		layout.PaddingBottom = node.PaddingBottom
		layout.SpacingH = node.ItemSpacing
		layout.SpacingV = node.ItemSpacing
		layout.LayoutAlign = node.LayoutAlign
		layout.LayoutGrow = node.LayoutGrow
	}

	// 境界ボックスからサイズを設定
	if node.AbsoluteBoundingBox != nil {
		layout.Width = node.AbsoluteBoundingBox.Width
		layout.Height = node.AbsoluteBoundingBox.Height
	}

	// レイアウト情報があれば追加
	if layout.LayoutMode != "" || layout.Width > 0 || layout.Height > 0 {
		layoutID := generateVarID("layout")
		globalVars.Styles[layoutID] = layout
		simplified.Layout = layoutID
	}

	// 不透明度の処理
	if node.Opacity != 0 && node.Opacity != 1 {
		simplified.Opacity = node.Opacity
	}

	// 角丸の処理
	if node.CornerRadius > 0 {
		simplified.BorderRadius = fmt.Sprintf("%.2fpx", node.CornerRadius)
	} else if len(node.RectangleCornerRadii) == 4 {
		simplified.BorderRadius = fmt.Sprintf("%.2fpx %.2fpx %.2fpx %.2fpx",
			node.RectangleCornerRadii[0], node.RectangleCornerRadii[1],
			node.RectangleCornerRadii[2], node.RectangleCornerRadii[3])
	}

	// 子ノードの処理
	if len(node.Children) > 0 {
		simplified.Children = make([]SimplifiedNode, 0, len(node.Children))
		for _, child := range node.Children {
			if isVisible(child) {
				childNode := parseNode(globalVars, child, node)
				if childNode.ID != "" {
					simplified.Children = append(simplified.Children, childNode)
				}
			}
		}
	}

	// VECTORタイプをIMAGE-SVGに変換
	if node.Type == "VECTOR" {
		simplified.Type = "IMAGE-SVG"
	}

	return simplified
}

// parseFill はFigmaの塗りつぶしを簡略化された構造に変換します
func parseFill(fill FigmaFill) SimplifiedFill {
	simplified := SimplifiedFill{
		Type:    fill.Type,
		Opacity: fill.Opacity,
	}

	// 色の処理
	if fill.Color != nil {
		r, g, b := int(fill.Color.R*255), int(fill.Color.G*255), int(fill.Color.B*255)
		a := fill.Color.A
		if a == 0 {
			a = 1
		}

		// 16進数形式
		simplified.Hex = fmt.Sprintf("#%02X%02X%02X", r, g, b)

		// RGBA形式
		simplified.RGBA = fmt.Sprintf("rgba(%d, %d, %d, %.2f)", r, g, b, a)
	}

	// 画像参照の処理
	if fill.ImageRef != "" {
		simplified.ImageRef = fill.ImageRef
		simplified.ScaleMode = fill.ScaleMode
	}

	// グラデーションの処理
	if len(fill.GradientHandlePositions) > 0 {
		simplified.GradientHandlePositions = fill.GradientHandlePositions
	}

	if len(fill.GradientStops) > 0 {
		simplified.GradientStops = make([]GradientStop, len(fill.GradientStops))
		for i, stop := range fill.GradientStops {
			r, g, b := int(stop.Color.R*255), int(stop.Color.G*255), int(stop.Color.B*255)
			hex := fmt.Sprintf("#%02X%02X%02X", r, g, b)

			simplified.GradientStops[i] = GradientStop{
				Position: stop.Position,
				Color: ColorValue{
					Hex:     hex,
					Opacity: stop.Color.A,
				},
			}
		}
	}

	return simplified
}

// isVisible はノードが表示可能かどうかを判定します
func isVisible(node FigmaDocumentNode) bool {
	return node.Visible == nil || *node.Visible
}

// generateVarID は変数IDを生成します
func generateVarID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}

// removeEmptyKeys は空のキーを削除します
func removeEmptyKeys(node SimplifiedNode) SimplifiedNode {
	// 実装は省略（必要に応じて実装）
	return node
}
