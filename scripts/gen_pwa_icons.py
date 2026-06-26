from pathlib import Path

from PIL import Image

STATIC = Path("static")
ICONS = STATIC / "icons"
FISH_SRC = STATIC / "images" / "sumit-fish.png"
KURENAI = (192, 20, 60, 255)
BG = (10, 10, 10, 255)


def load_fish() -> Image.Image:
    im = Image.open(FISH_SRC).convert("RGBA")
    bbox = im.getbbox()
    if not bbox:
        raise SystemExit(f"empty fish image: {FISH_SRC}")
    return im.crop(bbox)


def make_square_icon(
    fish: Image.Image,
    size: int,
    padding_ratio: float = 0.0,
    *,
    cover: bool = False,
    zoom: float = 1.0,
) -> Image.Image:
    canvas = Image.new("RGBA", (size, size), BG)
    fw, fh = fish.size
    pad = int(size * padding_ratio)
    avail = max(1, size - 2 * pad)

    if cover:
        scale = max(avail / fw, avail / fh) * zoom
        nw = max(1, int(fw * scale))
        nh = max(1, int(fh * scale))
        resized = fish.resize((nw, nh), Image.Resampling.LANCZOS)
        ox = (size - nw) // 2
        oy = (size - nh) // 2
        layer = Image.new("RGBA", (size, size), BG)
        layer.paste(resized, (ox, oy), resized)
        return layer

    scale = min(avail / fw, avail / fh)
    nw = max(1, int(fw * scale))
    nh = max(1, int(fh * scale))
    resized = fish.resize((nw, nh), Image.Resampling.LANCZOS)
    ox = (size - nw) // 2
    oy = (size - nh) // 2
    canvas.paste(resized, (ox, oy), resized)
    return canvas


def main() -> None:
    fish = load_fish()
    ICONS.mkdir(parents=True, exist_ok=True)

    make_square_icon(fish, 16, cover=True, zoom=1.12).save(ICONS / "favicon-16.png")
    make_square_icon(fish, 32, cover=True, zoom=1.10).save(ICONS / "favicon-32.png")
    make_square_icon(fish, 192, 0.08).save(ICONS / "pwa-192.png")
    make_square_icon(fish, 512, 0.10).save(ICONS / "pwa-512.png")


if __name__ == "__main__":
    main()
