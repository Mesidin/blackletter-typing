#!/usr/bin/env python3
"""Convert ImageMagick gray: bytes to Unicode braille art."""

import sys

DOTS = [
    (0, 0, 0x01),
    (0, 1, 0x02),
    (0, 2, 0x04),
    (1, 0, 0x08),
    (1, 1, 0x10),
    (1, 2, 0x20),
    (0, 3, 0x40),
    (1, 3, 0x80),
]
BLANK = "\u2800"


def pad(pixels, w, h):
    nw = w + (w % 2)
    nh = h + ((4 - (h % 4)) % 4)
    if nw == w and nh == h:
        return pixels, w, h
    out = bytearray(nw * nh)
    out[:] = b"\xff" * (nw * nh)
    for y in range(h):
        out[y * nw : y * nw + w] = pixels[y * w : y * w + w]
    return bytes(out), nw, nh


def to_braille(pixels, w, h, threshold=128):
    pixels, w, h = pad(pixels, w, h)
    rows = []
    for y in range(0, h, 4):
        chars = []
        for x in range(0, w, 2):
            bits = 0
            for dx, dy, mask in DOTS:
                px, py = x + dx, y + dy
                if pixels[py * w + px] < threshold:
                    bits |= mask
            chars.append(chr(0x2800 + bits))
        rows.append(chars)
    # trim empty rows
    while rows and all(c == BLANK for c in rows[0]):
        rows.pop(0)
    while rows and all(c == BLANK for c in rows[-1]):
        rows.pop()
    if not rows:
        return ""
    # trim empty columns
    left = 0
    width = len(rows[0])
    while left < width and all(r[left] == BLANK for r in rows):
        left += 1
    right = width
    while right > left and all(r[right - 1] == BLANK for r in rows):
        right -= 1
    return "\n".join("".join(r[left:right]).rstrip(BLANK) for r in rows)


def main():
    path, w, h = sys.argv[1], int(sys.argv[2]), int(sys.argv[3])
    thresh = int(sys.argv[4]) if len(sys.argv) > 4 else 128
    data = open(path, "rb").read()
    if len(data) != w * h:
        sys.exit(f"expected {w*h} bytes, got {len(data)}")
    print(to_braille(data, w, h, thresh))


if __name__ == "__main__":
    main()
