package valueobject

import "testing"

func TestSignalType_String(t *testing.T) {
	tests := []struct {
		name string
		st   SignalType
		want string
	}{
		{"STRONG_BUY", SignalStrongBuy, "STRONG_BUY"},
		{"BUY", SignalBuy, "BUY"},
		{"WATCH", SignalWatch, "WATCH"},
		{"AVOID", SignalAvoid, "AVOID"},
		{"STRONG_AVOID", SignalStrongAvoid, "STRONG_AVOID"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := string(tt.st); got != tt.want {
				t.Errorf("SignalType string = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFromScore(t *testing.T) {
	tests := []struct {
		name  string
		score int
		want  SignalType
	}{
		{"score_90", 90, SignalStrongBuy},
		{"score_85", 85, SignalStrongBuy},
		{"score_80", 80, SignalBuy},
		{"score_75", 75, SignalBuy},
		{"score_70", 70, SignalWatch},
		{"score_60", 60, SignalWatch},
		{"score_55", 55, SignalAvoid},
		{"score_50", 50, SignalAvoid},
		{"score_45", 45, SignalStrongAvoid},
		{"score_0", 0, SignalStrongAvoid},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FromScore(tt.score); got != tt.want {
				t.Errorf("FromScore(%v) = %v, want %v", tt.score, got, tt.want)
			}
		})
	}
}

func TestSignalType_IsValid(t *testing.T) {
	tests := []struct {
		name string
		st   SignalType
		want bool
	}{
		{"STRONG_BUY_valid", SignalStrongBuy, true},
		{"BUY_valid", SignalBuy, true},
		{"WATCH_valid", SignalWatch, true},
		{"AVOID_valid", SignalAvoid, true},
		{"STRONG_AVOID_valid", SignalStrongAvoid, true},
		{"invalid", SignalType("INVALID"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.st.IsValid(); got != tt.want {
				t.Errorf("SignalType.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSignalType_IsBullish(t *testing.T) {
	tests := []struct {
		name string
		st   SignalType
		want bool
	}{
		{"STRONG_BUY_bullish", SignalStrongBuy, true},
		{"BUY_bullish", SignalBuy, true},
		{"WATCH_not_bullish", SignalWatch, false},
		{"AVOID_not_bullish", SignalAvoid, false},
		{"STRONG_AVOID_not_bullish", SignalStrongAvoid, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.st.IsBullish(); got != tt.want {
				t.Errorf("SignalType.IsBullish() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSignalType_IsBearish(t *testing.T) {
	tests := []struct {
		name string
		st   SignalType
		want bool
	}{
		{"STRONG_BUY_not_bearish", SignalStrongBuy, false},
		{"BUY_not_bearish", SignalBuy, false},
		{"WATCH_not_bearish", SignalWatch, false},
		{"AVOID_bearish", SignalAvoid, true},
		{"STRONG_AVOID_bearish", SignalStrongAvoid, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.st.IsBearish(); got != tt.want {
				t.Errorf("SignalType.IsBearish() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRiskLevel_Constants(t *testing.T) {
	tests := []struct {
		name string
		rl   RiskLevel
		want string
	}{
		{"LOW", RiskLow, "LOW"},
		{"MEDIUM", RiskMedium, "MEDIUM"},
		{"HIGH", RiskHigh, "HIGH"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := string(tt.rl); got != tt.want {
				t.Errorf("RiskLevel string = %v, want %v", got, tt.want)
			}
		})
	}
}
