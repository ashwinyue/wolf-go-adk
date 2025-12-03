/*
 * Copyright 2025 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package utils

import (
	"fmt"
	"strings"
)

// Truncate 截断字符串（支持多字节字符）
func Truncate(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	runes := []rune(s)
	if len(runes) > maxLen {
		return string(runes[:maxLen]) + "..."
	}
	return s
}

// MajorityVote 多数投票
func MajorityVote(votes map[string]string) (string, string) {
	counts := make(map[string]int)
	for _, target := range votes {
		counts[target]++
	}

	var maxCount int
	var winner string
	var details []string

	for target, count := range counts {
		details = append(details, fmt.Sprintf("%s:%d", target, count))
		if count > maxCount {
			maxCount = count
			winner = target
		}
	}

	return winner, strings.Join(details, ", ")
}
