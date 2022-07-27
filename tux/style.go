package tux

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type StyleSheet struct {
	Styles map[string]StyleRule
	Tokens map[string]string
}

type StyleRule struct {
	Bold               bool
	Italic             bool
	Underline          bool
	Strikethrough      bool
	Blink              bool
	Faint              bool
	Foreground         string
	ForegroundInverted string
	Background         string
	BackgroundInverted string
	PaddingTop         int
	PaddingRight       int
	PaddingBottom      int
	PaddingLeft        int
	MarginTop          int
	MarginRight        int
	MarginBottom       int
	MarginLeft         int
}

func Render(styleSheet StyleSheet, class string, text string) string {
	return text
}

type StyleRenderer interface {
	Render(str string) string
}

func Renderer(styleRule StyleRule, tokens map[string]string) StyleRenderer {
	var renderer = lipgloss.NewStyle()
	renderer = renderer.Bold(styleRule.Bold)
	renderer = renderer.Italic(styleRule.Italic)
	renderer = renderer.Underline(styleRule.Underline)
	renderer = renderer.Strikethrough(styleRule.Strikethrough)
	renderer = renderer.Blink(styleRule.Blink)
	renderer = renderer.Faint(styleRule.Faint)
	if styleRule.Foreground != "" {
		renderer = renderer.Foreground(getColor(styleRule.Foreground, styleRule.ForegroundInverted, tokens))
	}
	if styleRule.Background != "" {
		renderer = renderer.Background(getColor(styleRule.Background, styleRule.BackgroundInverted, tokens))
	}
	renderer = renderer.PaddingTop(styleRule.PaddingTop)
	renderer = renderer.PaddingRight(styleRule.PaddingRight)
	renderer = renderer.PaddingBottom(styleRule.PaddingBottom)
	renderer = renderer.PaddingLeft(styleRule.PaddingLeft)
	renderer = renderer.MarginTop(styleRule.MarginTop)
	renderer = renderer.MarginRight(styleRule.MarginRight)
	renderer = renderer.MarginBottom(styleRule.MarginBottom)
	renderer = renderer.MarginLeft(styleRule.MarginLeft)
	return renderer
}

func getColor(token string, invertedToken string, tokens map[string]string) lipgloss.TerminalColor {
	color := resolveToken(token, tokens)
	invertedColor := resolveToken(invertedToken, tokens)

	if invertedColor == "" {
		return lipgloss.Color(color)
	} else {
		return lipgloss.AdaptiveColor{
			Dark:  color,
			Light: invertedColor,
		}
	}
}

func resolveToken(token string, tokens map[string]string) string {
	if strings.HasPrefix(token, "$") == true {
		resolved, ok := tokens[token]
		if ok == true {
			return resolved
		}
		return ANSI_COLORS[token]
	}
	return token
}

func StyleFunc(styleSheet StyleSheet) func(class string, text string) string {
	return func(class string, text string) string {
		styleRule, exists := styleSheet.Styles[class]
		// Return the text as is if the class is not found.
		if !exists {
			return text
		}
		result := Renderer(styleRule, styleSheet.Tokens).Render(text)
		return result
	}
}

// TODO: Add list of default ANSI named colors
var ANSI_COLORS = map[string]string{}