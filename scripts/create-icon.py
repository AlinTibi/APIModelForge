"""Rebuild the committed icon assets with Pillow; not required to build the app."""
from pathlib import Path
from PIL import Image, ImageDraw

root = Path(__file__).resolve().parents[1]
scale = 4
image = Image.new("RGBA", (256 * scale, 256 * scale))
draw = ImageDraw.Draw(image)
def box(bounds, radius, fill):
    draw.rounded_rectangle(tuple(int(n * scale) for n in bounds), radius=radius * scale, fill=fill)
def polygon(points, fill):
    draw.polygon([(x * scale, y * scale) for x, y in points], fill=fill)

box((8, 8, 248, 248), 48, "#142533")
brace = [(82,54),(65,54),(52,67),(52,106),(34,120),(34,136),(52,150),(52,189),(65,202),(82,202),(82,181),(73,181),(73,145),(57,128),(73,111),(73,75),(82,75)]
polygon(brace, "#7CE5C5")
polygon([(256-x,y) for x,y in brace], "#7CE5C5")
for y,width,color in [(76,54,"#7CE5C5"),(115,70,"#58C8DF"),(154,54,"#7CE5C5")]:
    box((128-width/2,y,128+width/2,y+26),6,color)
image = image.resize((256,256), Image.Resampling.LANCZOS)
image.save(root / "build/appicon.png")
image.save(root / "build/windows/icon.ico", sizes=[(16,16),(24,24),(32,32),(48,48),(64,64),(128,128),(256,256)])
points = " ".join(f"{x},{y}" for x,y in brace)
mirror = " ".join(f"{256-x},{y}" for x,y in brace)
svg = f'''<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 256 256">
  <rect x="8" y="8" width="240" height="240" rx="48" fill="#142533"/>
  <polygon points="{points}" fill="#7ce5c5"/>
  <polygon points="{mirror}" fill="#7ce5c5"/>
  <rect x="101" y="76" width="54" height="26" rx="6" fill="#7ce5c5"/>
  <rect x="93" y="115" width="70" height="26" rx="6" fill="#58c8df"/>
  <rect x="101" y="154" width="54" height="26" rx="6" fill="#7ce5c5"/>
</svg>
'''
(root / "build/appicon.svg").write_text(svg, encoding="utf-8")
