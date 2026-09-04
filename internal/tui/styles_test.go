package tui

import "testing"

func TestBuildStylesRejectsInvalidToken(t *testing.T) {
	_, err := BuildStyles(ThemeConfig{ActiveBorderColor: []string{"bogus"}})
	if err == nil {
		t.Fatal("expected invalid theme token error")
	}
}

func TestBuildStylesAllowsThemeOverride(t *testing.T) {
	theme := ThemeConfig{
		ActiveBorderColor: []string{"yellow", "bold"},
		StatusModeColor:   []string{"cyan", "bold"},
		Modal:             ThemeModalConfig{BorderColor: "#ffffff"},
	}
	styles, err := BuildStyles(theme)
	if err != nil {
		t.Fatal(err)
	}
	merged := MergeTheme(DefaultThemeConfig(), theme)
	if !styles.ActiveBorder.GetBold() {
		t.Fatal("expected active border to preserve bold attribute")
	}
	if merged.Modal.BorderColor != "#ffffff" {
		t.Fatalf("modal border color = %q, want %q", merged.Modal.BorderColor, "#ffffff")
	}
	if !styles.StatusMode.GetBold() {
		t.Fatal("expected status mode to preserve bold attribute")
	}
}

func TestValidateUIConfigRejectsBadBorderStyle(t *testing.T) {
	err := ValidateUIConfig(UIConfig{BorderStyle: "weird", SplitRatio: 0.45, TwoPaneMinWidth: 110})
	if err == nil {
		t.Fatal("expected invalid ui border style error")
	}
}
