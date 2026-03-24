package barcode

import (
	"fmt"
	"image"
	"image/color"
	"strings"
)

// ── MaxiCode encoding (ISO/IEC 16023) ─────────────────────────────────────────
//
// Ported from C# FastReport.Base/Barcode/BarcodeMaxiCode.cs (MaxiCodeImpl class).

// maxiCodeSet maps each ASCII byte (0–255) to its Code Set:
// 1=Set A, 2=Set B, 3=Set C, 4=Set D, 5=Set E, 0=ambiguous (fits multiple sets).
// Source: ISO/IEC 16023 Appendix A, C# MAXICODE_SET table.
var maxiCodeSet = [256]int{
	5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 0, 5, 5, 5, 5, 5, 5,
	5, 5, 5, 5, 5, 5, 5, 5, 0, 0, 0, 5, 0, 2, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 0, 1, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 2,
	2, 2, 2, 2, 2, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 2, 2, 2, 2, 2, 2, 2, 2, 2,
	2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2,
	2, 2, 2, 2, 2, 2, 2, 2, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 4, 4,
	4, 4, 4, 4, 4, 4, 4, 4, 4, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5,
	5, 4, 5, 5, 5, 5, 5, 5, 4, 5, 3, 4, 3, 5, 5, 4, 4, 3, 3, 3,
	4, 3, 5, 4, 4, 3, 3, 4, 3, 3, 3, 4, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4,
	4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4,
}

// maxiCodeSymbolChar maps each ASCII byte (0–255) to its symbol value within
// its code set. Source: ISO/IEC 16023 Appendix A, C# MAXICODE_SYMBOL_CHAR table.
var maxiCodeSymbolChar = [256]int{
	0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19,
	20, 21, 22, 23, 24, 25, 26, 30, 28, 29, 30, 35, 32, 53, 34, 35, 36, 37, 38, 39,
	40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 37,
	38, 39, 40, 41, 52, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
	16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 42, 43, 44, 45, 46, 0, 1, 2, 3,
	4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23,
	24, 25, 26, 32, 54, 34, 35, 36, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 47, 48,
	49, 50, 51, 52, 53, 54, 55, 56, 57, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 36,
	37, 37, 38, 39, 40, 41, 42, 43, 38, 44, 37, 39, 38, 45, 46, 40, 41, 39, 40, 41,
	42, 42, 47, 43, 44, 43, 44, 45, 45, 46, 47, 46, 0, 1, 2, 3, 4, 5, 6, 7,
	8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 32,
	33, 34, 35, 36, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
	16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 32, 33, 34, 35, 36,
}

// maxiCodeGrid maps each cell (row*30+col) in the 33×30 hexagonal grid to a
// bit position index. bit_pos = grid_val-1; codeword = bit_pos/6; bit = bit_pos%6.
// A value of 0 means the cell is part of the bullseye / unused.
// Source: ISO/IEC 16023 Figure 5, C# MAXICODE_GRID table.
var maxiCodeGrid = [33 * 30]int{
	122, 121, 128, 127, 134, 133, 140, 139, 146, 145, 152, 151, 158, 157, 164, 163, 170, 169, 176, 175, 182, 181, 188, 187, 194, 193, 200, 199, 0, 0,
	124, 123, 130, 129, 136, 135, 142, 141, 148, 147, 154, 153, 160, 159, 166, 165, 172, 171, 178, 177, 184, 183, 190, 189, 196, 195, 202, 201, 817, 0,
	126, 125, 132, 131, 138, 137, 144, 143, 150, 149, 156, 155, 162, 161, 168, 167, 174, 173, 180, 179, 186, 185, 192, 191, 198, 197, 204, 203, 819, 818,
	284, 283, 278, 277, 272, 271, 266, 265, 260, 259, 254, 253, 248, 247, 242, 241, 236, 235, 230, 229, 224, 223, 218, 217, 212, 211, 206, 205, 820, 0,
	286, 285, 280, 279, 274, 273, 268, 267, 262, 261, 256, 255, 250, 249, 244, 243, 238, 237, 232, 231, 226, 225, 220, 219, 214, 213, 208, 207, 822, 821,
	288, 287, 282, 281, 276, 275, 270, 269, 264, 263, 258, 257, 252, 251, 246, 245, 240, 239, 234, 233, 228, 227, 222, 221, 216, 215, 210, 209, 823, 0,
	290, 289, 296, 295, 302, 301, 308, 307, 314, 313, 320, 319, 326, 325, 332, 331, 338, 337, 344, 343, 350, 349, 356, 355, 362, 361, 368, 367, 825, 824,
	292, 291, 298, 297, 304, 303, 310, 309, 316, 315, 322, 321, 328, 327, 334, 333, 340, 339, 346, 345, 352, 351, 358, 357, 364, 363, 370, 369, 826, 0,
	294, 293, 300, 299, 306, 305, 312, 311, 318, 317, 324, 323, 330, 329, 336, 335, 342, 341, 348, 347, 354, 353, 360, 359, 366, 365, 372, 371, 828, 827,
	410, 409, 404, 403, 398, 397, 392, 391, 80, 79, 0, 0, 14, 13, 38, 37, 3, 0, 45, 44, 110, 109, 386, 385, 380, 379, 374, 373, 829, 0,
	412, 411, 406, 405, 400, 399, 394, 393, 82, 81, 41, 0, 16, 15, 40, 39, 4, 0, 0, 46, 112, 111, 388, 387, 382, 381, 376, 375, 831, 830,
	414, 413, 408, 407, 402, 401, 396, 395, 84, 83, 42, 0, 0, 0, 0, 0, 6, 5, 48, 47, 114, 113, 390, 389, 384, 383, 378, 377, 832, 0,
	416, 415, 422, 421, 428, 427, 104, 103, 56, 55, 17, 0, 0, 0, 0, 0, 0, 0, 21, 20, 86, 85, 434, 433, 440, 439, 446, 445, 834, 833,
	418, 417, 424, 423, 430, 429, 106, 105, 58, 57, 0, 0, 0, 0, 0, 0, 0, 0, 23, 22, 88, 87, 436, 435, 442, 441, 448, 447, 835, 0,
	420, 419, 426, 425, 432, 431, 108, 107, 60, 59, 0, 0, 0, 0, 0, 0, 0, 0, 0, 24, 90, 89, 438, 437, 444, 443, 450, 449, 837, 836,
	482, 481, 476, 475, 470, 469, 49, 0, 31, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 54, 53, 464, 463, 458, 457, 452, 451, 838, 0,
	484, 483, 478, 477, 472, 471, 50, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 466, 465, 460, 459, 454, 453, 840, 839,
	486, 485, 480, 479, 474, 473, 52, 51, 32, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2, 0, 43, 468, 467, 462, 461, 456, 455, 841, 0,
	488, 487, 494, 493, 500, 499, 98, 97, 62, 61, 0, 0, 0, 0, 0, 0, 0, 0, 0, 27, 92, 91, 506, 505, 512, 511, 518, 517, 843, 842,
	490, 489, 496, 495, 502, 501, 100, 99, 64, 63, 0, 0, 0, 0, 0, 0, 0, 0, 29, 28, 94, 93, 508, 507, 514, 513, 520, 519, 844, 0,
	492, 491, 498, 497, 504, 503, 102, 101, 66, 65, 18, 0, 0, 0, 0, 0, 0, 0, 19, 30, 96, 95, 510, 509, 516, 515, 522, 521, 846, 845,
	560, 559, 554, 553, 548, 547, 542, 541, 74, 73, 33, 0, 0, 0, 0, 0, 0, 11, 68, 67, 116, 115, 536, 535, 530, 529, 524, 523, 847, 0,
	562, 561, 556, 555, 550, 549, 544, 543, 76, 75, 0, 0, 8, 7, 36, 35, 12, 0, 70, 69, 118, 117, 538, 537, 532, 531, 526, 525, 849, 848,
	564, 563, 558, 557, 552, 551, 546, 545, 78, 77, 0, 34, 10, 9, 26, 25, 0, 0, 72, 71, 120, 119, 540, 539, 534, 533, 528, 527, 850, 0,
	566, 565, 572, 571, 578, 577, 584, 583, 590, 589, 596, 595, 602, 601, 608, 607, 614, 613, 620, 619, 626, 625, 632, 631, 638, 637, 644, 643, 852, 851,
	568, 567, 574, 573, 580, 579, 586, 585, 592, 591, 598, 597, 604, 603, 610, 609, 616, 615, 622, 621, 628, 627, 634, 633, 640, 639, 646, 645, 853, 0,
	570, 569, 576, 575, 582, 581, 588, 587, 594, 593, 600, 599, 606, 605, 612, 611, 618, 617, 624, 623, 630, 629, 636, 635, 642, 641, 648, 647, 855, 854,
	728, 727, 722, 721, 716, 715, 710, 709, 704, 703, 698, 697, 692, 691, 686, 685, 680, 679, 674, 673, 668, 667, 662, 661, 656, 655, 650, 649, 856, 0,
	730, 729, 724, 723, 718, 717, 712, 711, 706, 705, 700, 699, 694, 693, 688, 687, 682, 681, 676, 675, 670, 669, 664, 663, 658, 657, 652, 651, 858, 857,
	732, 731, 726, 725, 720, 719, 714, 713, 708, 707, 702, 701, 696, 695, 690, 689, 684, 683, 678, 677, 672, 671, 666, 665, 660, 659, 654, 653, 859, 0,
	734, 733, 740, 739, 746, 745, 752, 751, 758, 757, 764, 763, 770, 769, 776, 775, 782, 781, 788, 787, 794, 793, 800, 799, 806, 805, 812, 811, 861, 860,
	736, 735, 742, 741, 748, 747, 754, 753, 760, 759, 766, 765, 772, 771, 778, 777, 784, 783, 790, 789, 796, 795, 802, 801, 808, 807, 814, 813, 862, 0,
	738, 737, 744, 743, 750, 749, 756, 755, 762, 761, 768, 767, 774, 773, 780, 779, 786, 785, 792, 791, 798, 797, 804, 803, 810, 809, 816, 815, 864, 863,
}

// MaxiCodeComputeECC is exported for testing; it computes RS ECC for MaxiCode.
func MaxiCodeComputeECC(data []byte, eccLen int) []byte { return maxiCodeRS(data, eccLen) }

// maxiCodeGFTable holds pre-computed GF(64) log/antilog tables.
// Generator polynomial: x^6 + x + 1 (0x43).
var maxiCodeGFTable = func() struct {
	logt [64]int
	alog [63]int
} {
	var gf struct {
		logt [64]int
		alog [63]int
	}
	p := 1
	for v := 0; v < 63; v++ {
		gf.alog[v] = p
		gf.logt[p] = v
		p <<= 1
		if p&64 != 0 {
			p ^= 0x43
		}
	}
	return gf
}()

// maxiCodeRS computes Reed-Solomon ECC for MaxiCode over GF(64).
// Returns eccLen ECC codewords with index 0 = highest-degree check symbol.
func maxiCodeRS(data []byte, eccLen int) []byte {
	gf := &maxiCodeGFTable
	const logmod = 63

	rspoly := make([]int, eccLen+1)
	rspoly[0] = 1
	for i := 1; i <= eccLen; i++ {
		rspoly[i] = 1
		for k := i - 1; k > 0; k-- {
			if rspoly[k] != 0 {
				rspoly[k] = gf.alog[(gf.logt[rspoly[k]]+i)%logmod]
			}
			rspoly[k] ^= rspoly[k-1]
		}
		rspoly[0] = gf.alog[(gf.logt[rspoly[0]]+i)%logmod]
	}

	res := make([]int, eccLen)
	for _, d := range data {
		m := res[eccLen-1] ^ int(d)
		for k := eccLen - 1; k > 0; k-- {
			if m != 0 && rspoly[k] != 0 {
				res[k] = res[k-1] ^ gf.alog[(gf.logt[m]+gf.logt[rspoly[k]])%logmod]
			} else {
				res[k] = res[k-1]
			}
		}
		if m != 0 && rspoly[0] != 0 {
			res[0] = gf.alog[(gf.logt[m]+gf.logt[rspoly[0]])%logmod]
		} else {
			res[0] = 0
		}
	}

	ecc := make([]byte, eccLen)
	for i := 0; i < eccLen; i++ {
		ecc[i] = byte(res[eccLen-1-i])
	}
	return ecc
}

// maxiCodeRSInt is like maxiCodeRS but operates on int slices (used internally).
func maxiCodeRSInt(data []int, eccLen int) []int {
	tmp := make([]byte, len(data))
	for i, v := range data {
		tmp[i] = byte(v)
	}
	raw := maxiCodeRS(tmp, eccLen)
	out := make([]int, len(raw))
	for i, v := range raw {
		out[i] = int(v)
	}
	return out
}

// maxiCodeProcessText builds the set[] and character[] encoding arrays from
// input text using the ISO/IEC 16023 Appendix A lookup tables, then inserts
// latch/shift control codes for set transitions.
// Ported from C# MaxiCodeImpl.processText().
// Returns (set[144], character[144], textLength) or an error if too long.
func maxiCodeProcessText(text string, mode int) (set [144]int, character [144]int, length int, err error) {
	src := []byte(text)
	length = len(src)
	if length > 138 {
		// Truncate to maximum encodable length (matches C# behaviour of limiting input).
		src = src[:138]
		length = 138
	}

	for i := 0; i < 144; i++ {
		set[i] = -1
		character[i] = 0
	}

	// Initial assignment from lookup tables.
	for i := 0; i < length; i++ {
		b := int(src[i]) & 0xFF
		set[i] = maxiCodeSet[b]
		character[i] = maxiCodeSymbolChar[b]
	}

	// Resolve ambiguous characters (set == 0) using context.
	if set[0] == 0 {
		if character[0] == 13 {
			character[0] = 0
		}
		set[0] = 1
	}
	for i := 1; i < length; i++ {
		if set[i] == 0 {
			switch character[i] {
			case 13: // CR
				set[i] = bestSurroundingSet(set[:], i, length, 1, 5)
				if set[i] == 5 {
					character[i] = 13
				} else {
					character[i] = 0
				}
			case 28: // FS
				set[i] = bestSurroundingSet(set[:], i, length, 1, 2, 3, 4, 5)
				if set[i] == 5 {
					character[i] = 32
				}
			case 29: // GS
				set[i] = bestSurroundingSet(set[:], i, length, 1, 2, 3, 4, 5)
				if set[i] == 5 {
					character[i] = 33
				}
			case 30: // RS
				set[i] = bestSurroundingSet(set[:], i, length, 1, 2, 3, 4, 5)
				if set[i] == 5 {
					character[i] = 34
				}
			case 32: // Space
				set[i] = bestSurroundingSet(set[:], i, length, 1, 2, 3, 4, 5)
				if set[i] == 1 {
					character[i] = 32
				} else if set[i] == 2 {
					character[i] = 47
				} else {
					character[i] = 59
				}
			case 44: // Comma
				set[i] = bestSurroundingSet(set[:], i, length, 1, 2)
				if set[i] == 2 {
					character[i] = 48
				}
			case 46: // Full stop
				set[i] = bestSurroundingSet(set[:], i, length, 1, 2)
				if set[i] == 2 {
					character[i] = 49
				}
			case 47: // Slash
				set[i] = bestSurroundingSet(set[:], i, length, 1, 2)
				if set[i] == 2 {
					character[i] = 50
				}
			case 58: // Colon
				set[i] = bestSurroundingSet(set[:], i, length, 1, 2)
				if set[i] == 2 {
					character[i] = 51
				}
			}
		}
	}

	// Pad remaining entries.
	lastSet := 1
	if length > 0 {
		lastSet = set[length-1]
	}
	for i := length; i < 144; i++ {
		if lastSet == 2 {
			set[i] = 2
		} else {
			set[i] = 1
		}
		character[i] = 33
	}

	// Number compression: 9 consecutive Set A digits → 6 codewords.
	startJ := 0
	if mode == 2 || mode == 3 {
		startJ = 9
	}
	count := 0
	for i := startJ; i < 143; i++ {
		if set[i] == 1 && character[i] >= 48 && character[i] <= 57 {
			count++
		} else {
			count = 0
		}
		if count == 9 {
			set[i] = 6
			set[i-1] = 6
			set[i-2] = 6
			set[i-3] = 6
			set[i-4] = 6
			set[i-5] = 6
			set[i-6] = 6
			set[i-7] = 6
			set[i-8] = 6
			count = 0
		}
	}

	// Insert latch/shift characters between sets.
	currentSet := 1
	i := 0
	for i < 144 {
		if set[i] != currentSet && set[i] != 6 {
			switch set[i] {
			case 1: // need Set A while in Set B/C/D/E
				if i+1 < 144 && set[i+1] == 1 {
					if i+2 < 144 && set[i+2] == 1 {
						if i+3 < 144 && set[i+3] == 1 {
							// Latch A
							maxiInsert(set[:], character[:], i, 63)
							currentSet = 1
							length++
							i += 3
						} else {
							// 3 Shift A
							maxiInsert(set[:], character[:], i, 57)
							length++
							i += 2
						}
					} else {
						// 2 Shift A
						maxiInsert(set[:], character[:], i, 56)
						length++
						i++
					}
				} else {
					// Shift A
					maxiInsert(set[:], character[:], i, 59)
					length++
				}
			case 2: // need Set B while in Set A/C/D/E
				if i+1 < 144 && set[i+1] == 2 {
					// Latch B
					maxiInsert(set[:], character[:], i, 63)
					currentSet = 2
					length++
					i++
				} else {
					// Shift B
					maxiInsert(set[:], character[:], i, 59)
					length++
				}
			case 3:
				if i+3 < 144 && set[i+1] == 3 && set[i+2] == 3 && set[i+3] == 3 {
					// Lock In C (two 60s)
					maxiInsert(set[:], character[:], i, 60)
					maxiInsert(set[:], character[:], i, 60)
					currentSet = 3
					length++
					i += 3
				} else {
					// Shift C
					maxiInsert(set[:], character[:], i, 60)
					length++
				}
			case 4:
				if i+3 < 144 && set[i+1] == 4 && set[i+2] == 4 && set[i+3] == 4 {
					// Lock In D
					maxiInsert(set[:], character[:], i, 61)
					maxiInsert(set[:], character[:], i, 61)
					currentSet = 4
					length++
					i += 3
				} else {
					// Shift D
					maxiInsert(set[:], character[:], i, 61)
					length++
				}
			case 5:
				if i+3 < 144 && set[i+1] == 5 && set[i+2] == 5 && set[i+3] == 5 {
					// Lock In E
					maxiInsert(set[:], character[:], i, 62)
					maxiInsert(set[:], character[:], i, 62)
					currentSet = 5
					length++
					i += 3
				} else {
					// Shift E
					maxiInsert(set[:], character[:], i, 62)
					length++
				}
			}
			i++
		}
		i++
	}

	// Apply number compression values.
	i = 0
	for i < 144 {
		if set[i] == 6 {
			value := 0
			for j := 0; j < 9; j++ {
				value *= 10
				value += character[i+j] - '0'
			}
			character[i] = 31 // NS
			character[i+1] = (value & 0x3f000000) >> 24
			character[i+2] = (value & 0xfc0000) >> 18
			character[i+3] = (value & 0x3f000) >> 12
			character[i+4] = (value & 0xfc0) >> 6
			character[i+5] = value & 0x3f
			i += 6
			// shift remaining down by 3
			for j := i; j < 140; j++ {
				set[j] = set[j+3]
				character[j] = character[j+3]
			}
			length -= 3
		} else {
			i++
		}
	}

	return
}

// maxiInsert shifts set[pos..143] and character[pos..143] up by 1 and
// sets character[pos] = c (with set[pos] implicitly inherited from shift).
// Ported from C# MaxiCodeImpl.insert(position, c).
func maxiInsert(set []int, character []int, pos, c int) {
	for i := 143; i > pos; i-- {
		set[i] = set[i-1]
		character[i] = character[i-1]
	}
	character[pos] = c
}

// bestSurroundingSet picks the best code set for an ambiguous character at
// index i by looking at surrounding characters.
// Ported from C# MaxiCodeImpl.bestSurroundingSet().
func bestSurroundingSet(set []int, index, length int, valid ...int) int {
	contains := func(v int) bool {
		for _, x := range valid {
			if x == v {
				return true
			}
		}
		return false
	}
	opt1 := set[index-1]
	if index+1 < length {
		opt2 := set[index+1]
		if contains(opt1) && contains(opt2) {
			if opt1 < opt2 {
				return opt1
			}
			return opt2
		} else if contains(opt1) {
			return opt1
		} else if contains(opt2) {
			return opt2
		}
		return valid[0]
	}
	if contains(opt1) {
		return opt1
	}
	return valid[0]
}

// maxiCodeEncode encodes text as a 144-codeword MaxiCode symbol with RS ECC.
// Returns a 33×30 boolean grid (true = dark cell) plus orientation marks.
// Ported from C# MaxiCodeImpl.encode() and plotSymbol().
func maxiCodeEncode(text string, mode int) ([33][30]bool, error) {
	if mode < 2 || mode > 6 {
		mode = 4
	}

	secondaryMax := 84
	secondaryECMax := 40
	if mode == 5 {
		secondaryMax = 68
		secondaryECMax = 56
	}
	totalMax := secondaryMax + 10

	codewords := make([]int, 144)

	if mode == 2 || mode == 3 {
		// Parse the structured payload.
		postcode, countryStr, service, secondary := maxiCodeParseMode23Text(text)

		countryNum := 0
		for _, ch := range countryStr {
			if ch >= '0' && ch <= '9' {
				countryNum = countryNum*10 + int(ch-'0')
			}
		}

		if mode == 2 {
			for i := 0; i < len(postcode) && i < 9; i++ {
				if postcode[i] != ' ' && (postcode[i] < '0' || postcode[i] > '9') {
					mode = 3
					break
				}
			}
		}

		var primaryCW []int
		if mode == 2 {
			primaryCW = maxiCodeMode2PrimaryCodewordsInt(strings.TrimRight(postcode, " "), countryNum, service)
		} else {
			primaryCW = maxiCodeMode3PrimaryCodewordsInt(strings.TrimRight(postcode, " "), countryNum, service)
		}
		copy(codewords[:10], primaryCW)

		// Encode secondary.
		_, secChar, _, err := maxiCodeProcessText(secondary, mode)
		if err != nil {
			return [33][30]bool{}, err
		}
		for i := 0; i < secondaryMax; i++ {
			codewords[10+i] = secChar[i]
		}
	} else {
		// Modes 4/5/6: mode byte + text.
		codewords[0] = mode
		_, ch, _, err := maxiCodeProcessText(text, mode)
		if err != nil {
			return [33][30]bool{}, err
		}
		for i := 0; i < totalMax-1; i++ {
			codewords[1+i] = ch[i]
		}
	}

	// Truncate to max data size.
	data := codewords[:totalMax]

	// Primary ECC.
	primaryECC := maxiCodeRSInt(data[:10], 10)

	// Secondary ECC (interleaved odd/even).
	secondary := data[10:]
	half := len(secondary) / 2
	secOdd := make([]int, half)
	secEven := make([]int, half)
	for i, cw := range secondary {
		if i%2 == 1 {
			secOdd[(i-1)/2] = cw
		} else {
			secEven[i/2] = cw
		}
	}
	eccHalf := secondaryECMax / 2
	secECCOdd := maxiCodeRSInt(secOdd, eccHalf)
	secECCEven := maxiCodeRSInt(secEven, eccHalf)

	// Assemble final 144 codewords.
	out := make([]int, 144)
	copy(out[0:10], data[:10])
	copy(out[10:20], primaryECC)
	copy(out[20:20+secondaryMax], secondary)
	for i := 0; i < len(secECCOdd); i++ {
		out[20+secondaryMax+(2*i)+1] = secECCOdd[i]
	}
	for i := 0; i < len(secECCEven); i++ {
		out[20+secondaryMax+(2*i)] = secECCEven[i]
	}

	// Copy data into 33×30 symbol grid using MAXICODE_GRID.
	var grid [33][30]bool
	for row := 0; row < 33; row++ {
		for col := 0; col < 30; col++ {
			gv := maxiCodeGrid[row*30+col]
			if gv == 0 {
				continue
			}
			block := (gv + 5) / 6 // 1-based codeword index
			bit := (gv + 5) % 6   // bit index within codeword (0=MSB)
			if block > 0 && block <= 144 {
				cw := out[block-1]
				bitPattern := [6]int{
					(cw & 0x20) >> 5,
					(cw & 0x10) >> 4,
					(cw & 0x08) >> 3,
					(cw & 0x04) >> 2,
					(cw & 0x02) >> 1,
					cw & 0x01,
				}
				grid[row][col] = bitPattern[bit] != 0
			}
		}
	}

	// Orientation marks (ISO/IEC 16023 §7.7).
	grid[0][28] = true
	grid[0][29] = true
	grid[9][10] = true
	grid[9][11] = true
	grid[10][11] = true
	grid[15][7] = true
	grid[16][8] = true
	grid[16][20] = true
	grid[17][20] = true
	grid[22][10] = true
	grid[23][10] = true
	grid[22][17] = true
	grid[23][17] = true

	return grid, nil
}

// maxiCodeMode2PrimaryCodewordsInt returns 10 primary codewords for Mode 2.
func maxiCodeMode2PrimaryCodewordsInt(postcode string, country, service int) []int {
	for i := 0; i < len(postcode); i++ {
		if postcode[i] < '0' || postcode[i] > '9' {
			postcode = postcode[:i]
			break
		}
	}
	if len(postcode) == 0 {
		postcode = "0"
	}
	postcodeNum := 0
	for _, ch := range postcode {
		postcodeNum = postcodeNum*10 + int(ch-'0')
	}
	postcodeLen := len(postcode)
	p := make([]int, 10)
	p[0] = ((postcodeNum & 0x03) << 4) | 2
	p[1] = (postcodeNum & 0xfc) >> 2
	p[2] = (postcodeNum & 0x3f00) >> 8
	p[3] = (postcodeNum & 0xfc000) >> 14
	p[4] = (postcodeNum & 0x3f00000) >> 20
	p[5] = ((postcodeNum & 0x3c000000) >> 26) | ((postcodeLen & 0x3) << 4)
	p[6] = ((postcodeLen & 0x3c) >> 2) | ((country & 0x3) << 4)
	p[7] = (country & 0xfc) >> 2
	p[8] = ((country & 0x300) >> 8) | ((service & 0xf) << 2)
	p[9] = (service & 0x3f0) >> 4
	return p
}

// maxiCodeMode3PrimaryCodewordsInt returns 10 primary codewords for Mode 3.
func maxiCodeMode3PrimaryCodewordsInt(postcode string, country, service int) []int {
	for len(postcode) < 6 {
		postcode += " "
	}
	postcode = postcode[:6]
	upper := strings.ToUpper(postcode)
	nums := make([]int, 6)
	for i := 0; i < 6; i++ {
		ch := int(upper[i])
		if upper[i] >= 'A' && upper[i] <= 'Z' {
			ch -= 64
		}
		if ch == 27 || ch == 31 || ch == 33 || ch >= 59 {
			ch = 32
		}
		nums[i] = ch
	}
	p := make([]int, 10)
	p[0] = ((nums[5] & 0x03) << 4) | 3
	p[1] = ((nums[4] & 0x03) << 4) | ((nums[5] & 0x3c) >> 2)
	p[2] = ((nums[3] & 0x03) << 4) | ((nums[4] & 0x3c) >> 2)
	p[3] = ((nums[2] & 0x03) << 4) | ((nums[3] & 0x3c) >> 2)
	p[4] = ((nums[1] & 0x03) << 4) | ((nums[2] & 0x3c) >> 2)
	p[5] = ((nums[0] & 0x03) << 4) | ((nums[1] & 0x3c) >> 2)
	p[6] = ((nums[0] & 0x3c) >> 2) | ((country & 0x3) << 4)
	p[7] = (country & 0xfc) >> 2
	p[8] = ((country & 0x300) >> 8) | ((service & 0xf) << 2)
	p[9] = (service & 0x3f0) >> 4
	return p
}

// maxiCodeParseMode23Text parses a Mode 2/3 payload string.
// Format: postal(9) + country(3) + service(2) + GS(1) + secondary
func maxiCodeParseMode23Text(text string) (postcode, country string, service int, secondary string) {
	for len(text) < 14 {
		text += " "
	}
	postcode = text[:9]
	country = text[9:12]
	svc := text[12:14]
	for _, ch := range svc {
		if ch >= '0' && ch <= '9' {
			service = service*10 + int(ch-'0')
		}
	}
	if len(text) > 14 && text[14] == 0x1D {
		secondary = text[15:]
	} else if len(text) > 14 {
		secondary = text[14:]
	}
	return
}

// MaxiCodeMode2Payload builds a structured carrier message for Mode 2.
func MaxiCodeMode2Payload(zipCode, countryCode, serviceClass, secondary string) string {
	zip := padRight(zipCode, 9, ' ')
	country := padRight(countryCode, 3, ' ')
	svc := padRight(serviceClass, 2, ' ')
	return zip + country + svc + "\x1d" + secondary
}

// MaxiCodeMode3Payload builds a structured carrier message for Mode 3.
func MaxiCodeMode3Payload(zipCode, countryCode, serviceClass, secondary string) string {
	return MaxiCodeMode2Payload(zipCode, countryCode, serviceClass, secondary)
}

func padRight(s string, n int, pad byte) string {
	if len(s) >= n {
		return s[:n]
	}
	return s + strings.Repeat(string(pad), n-len(s))
}

// Encode validates the mode and stores encoded text for Render.
func (b *MaxiCodeBarcode) Encode(text string) error {
	if b.Mode < 2 || b.Mode > 6 {
		return fmt.Errorf("maxicode: invalid mode %d (must be 2–6)", b.Mode)
	}
	b.encodedText = text
	return nil
}

// Render produces a MaxiCode image.
// Ported from C# MaxiCodeImpl.plotSymbol() + BarcodeMaxiCode.Draw2DBarcode().
//
// Cell coordinates follow the C# formula:
//   x = (2.46 * col) + 1.23  (+ 1.23 for odd rows)
//   y = (2.135 * row) + 1.43
//
// The overall symbol spans ~0..76 in x and ~0..72 in y.
// We scale these to the requested width/height.
func (b *MaxiCodeBarcode) Render(width, height int) (image.Image, error) {
	if b.encodedText == "" {
		return nil, fmt.Errorf("maxicode: not encoded")
	}
	if width <= 0 {
		width = 100
	}
	if height <= 0 {
		height = 100
	}

	grid, err := maxiCodeEncode(b.encodedText, b.Mode)
	if err != nil {
		return nil, err
	}
	return maxiCodeRender(grid, width, height), nil
}

// maxiCodeRender draws a MaxiCode symbol from a populated 33×30 grid.
// Coordinates match C# plotSymbol (x=2.46*col+1.23, y=2.135*row+1.43).
// FieldSizeFactor=2.47 (C# constant) scales the symbol to the target size.
func maxiCodeRender(grid [33][30]bool, width, height int) image.Image {
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	white := color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	black := color.NRGBA{A: 255}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.SetNRGBA(x, y, white)
		}
	}

	// Symbol coordinate space: x in [0..~76], y in [0..~72].
	// C# FieldSizeFactor=2.47 is applied to each coordinate.
	// We compute kx, ky to map the full symbol into (width, height).
	const symW = 2.47 * (2.46*29 + 1.23 + 1.25) // approx symbol width
	const symH = 2.47 * (2.135*32 + 1.43 + 1.0)  // approx symbol height
	kx := float64(width) / symW
	ky := float64(height) / symH

	// Draw data hexagons.
	for row := 0; row < 33; row++ {
		for col := 0; col < 30; col++ {
			if !grid[row][col] {
				continue
			}
			cx := (2.46*float64(col) + 1.23) * 2.47
			if row&1 != 0 {
				cx += 1.23 * 2.47
			}
			cy := (2.135*float64(row) + 1.43) * 2.47

			drawHex(img, cx*kx, cy*ky, 1.23*2.47*kx, 1.0675*2.47*ky, black)
		}
	}

	// Draw bullseye: 3 concentric ring outlines.
	// C# radii: {9.91, 6.16, 2.37}, centre: (35.76, 35.60).
	cx := 35.76 * 2.47 * kx
	cy := 35.60 * 2.47 * ky
	radii := [3]float64{9.91, 6.16, 2.37}
	colors := [3]color.NRGBA{black, white, black}
	for i := 2; i >= 0; i-- {
		r := radii[i] * 2.47
		drawFilledCircle(img, cx, cy, r*kx, r*ky, colors[i])
	}

	return img
}

// drawFilledCircle fills an ellipse in the image.
func drawFilledCircle(img *image.NRGBA, cx, cy, rx, ry float64, c color.NRGBA) {
	bounds := img.Bounds()
	x0 := int(cx - rx)
	y0 := int(cy - ry)
	x1 := int(cx+rx) + 1
	y1 := int(cy+ry) + 1
	for py := y0; py <= y1; py++ {
		if py < bounds.Min.Y || py >= bounds.Max.Y {
			continue
		}
		for px := x0; px <= x1; px++ {
			if px < bounds.Min.X || px >= bounds.Max.X {
				continue
			}
			dx := float64(px) - cx
			dy := float64(py) - cy
			if (dx*dx)/(rx*rx)+(dy*dy)/(ry*ry) <= 1.0 {
				img.SetNRGBA(px, py, c)
			}
		}
	}
}

// drawHex draws a filled hexagon approximated as an ellipse.
func drawHex(img *image.NRGBA, cx, cy, rx, ry float64, c color.NRGBA) {
	drawFilledCircle(img, cx, cy, rx, ry, c)
}


