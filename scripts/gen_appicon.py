from PIL import Image, ImageDraw, ImageFont

size = 1024
img = Image.new("RGBA", (size, size), (0, 0, 0, 0))
grad = Image.new("RGBA", (size, size), (0, 0, 0, 0))
gdraw = ImageDraw.Draw(grad)

light = (52, 199, 89, 255)
base = (16, 124, 65, 255)
for y in range(size):
    t = y / (size - 1)
    r = int(light[0] * (1 - t) + base[0] * t)
    g = int(light[1] * (1 - t) + base[1] * t)
    b = int(light[2] * (1 - t) + base[2] * t)
    gdraw.line([(0, y), (size, y)], fill=(r, g, b, 255))

mask = Image.new("L", (size, size), 0)
mdraw = ImageDraw.Draw(mask)
margin = 64
radius = 220
mdraw.rounded_rectangle(
    [margin, margin, size - margin, size - margin], radius=radius, fill=255
)
img = Image.composite(grad, img, mask)
draw = ImageDraw.Draw(img)

# White sheet
pad = 220
sheet_r = 48
top = pad - 40
bot = size - pad + 40
draw.rounded_rectangle([pad, top, size - pad, bot], radius=sheet_r, fill=(255, 255, 255, 255))

# Header band
hdr_bot = top + 120
draw.rounded_rectangle([pad, top, size - pad, hdr_bot], radius=sheet_r, fill=(210, 242, 222, 255))
draw.rectangle([pad, hdr_bot - sheet_r, size - pad, hdr_bot], fill=(210, 242, 222, 255))

# Grid
x0, y0 = pad + 36, hdr_bot + 28
x1, y1 = size - pad - 36, bot - 28
cols, rows = 3, 4
line = (16, 124, 65, 170)
for i in range(1, cols):
    x = x0 + (x1 - x0) * i / cols
    draw.line([(x, hdr_bot), (x, y1)], fill=line, width=14)
for j in range(1, rows):
    y = y0 + (y1 - y0) * j / rows
    draw.line([(x0, y), (x1, y)], fill=line, width=12)

# Sheet outline
draw.rounded_rectangle(
    [pad, top, size - pad, bot], radius=sheet_r, outline=(255, 255, 255, 235), width=10
)

# Excel-like X in header
font_path = "C:/Windows/Fonts/segoeuib.ttf"
try:
    font = ImageFont.truetype(font_path, 180)
except Exception:
    font = ImageFont.load_default()
bbox = draw.textbbox((0, 0), "X", font=font)
tw, th = bbox[2] - bbox[0], bbox[3] - bbox[1]
tx = (size - tw) // 2 - bbox[0]
ty = top + (hdr_bot - top - th) // 2 - bbox[1]
draw.text((tx, ty), "X", font=font, fill=(16, 124, 65, 255))

out = r"C:\Users\MINFU LIAO\XiaomiMiMoProjects\.mimo-sessions\Excel tool\ExcelTools\build\appicon.png"
img.save(out, "PNG")
print("saved", out, img.size)
