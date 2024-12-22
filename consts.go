// SPDX-FileCopyrightText: 2024 Jeff Mitchell <jeffrey.mitchell@gmail.com>
// SPDX-License-Identifier: APL-2.0

package twinkly

type Code int

const (
	CodeOk                   Code = 1000
	CodeError                Code = 1001
	CodeInvalidArgumentValue Code = 1101
	CodeInvalidArgumentKey   Code = 1105
	CodeDuplicateUniqueId    Code = 1106
)

type LedOperationMode string

const (
	LedOperationModeOff      LedOperationMode = "off"
	LedOperationModeColor    LedOperationMode = "color"
	LedOperationModeDemo     LedOperationMode = "demo"
	LedOperationModeEffect   LedOperationMode = "effect"
	LedOperationModeMovie    LedOperationMode = "movie"
	LedOperationModePlaylist LedOperationMode = "playlist"
	LedOperationModeRt       LedOperationMode = "rt"
)
