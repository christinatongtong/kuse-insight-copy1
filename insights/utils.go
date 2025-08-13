package insights

import (
	"fmt"
	"strconv"
)

const (
	CONFIDENCE_EDGE = 0.6
)

func extractHighConfidenceValue(candidates *Candidates) string {
	if candidates == nil || len(candidates.Candidates) == 0 {
		return ""
	}

	for _, candidate := range candidates.Candidates {
		if candidate == nil {
			continue
		}

		// 转换confidence为float64进行比较
		var confidence float64
		switch v := candidate.Confidence.(type) {
		case float64:
			confidence = v
		case float32:
			confidence = float64(v)
		case int:
			confidence = float64(v)
		case string:
			if parsed, err := strconv.ParseFloat(v, 64); err == nil {
				confidence = parsed
			} else {
				continue // 如果无法解析，跳过这个候选项
			}
		default:
			continue
		}

		if confidence > CONFIDENCE_EDGE {
			switch v := candidate.Value.(type) {
			case string:
				return v
			case int:
				return strconv.Itoa(v)
			case float64:
				return strconv.FormatFloat(v, 'f', -1, 64)
			case float32:
				return strconv.FormatFloat(float64(v), 'f', -1, 32)
			case bool:
				return strconv.FormatBool(v)
			default:
				return fmt.Sprintf("%v", v)
			}
		}
	}

	return ""
}
