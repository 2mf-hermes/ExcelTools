"""Build a multi-size Windows .ico from build/appicon.png."""
from PIL import Image
import os

root = r"C:\Users\MINFU LIAO\XiaomiMiMoProjects\.mimo-sessions\Excel tool\ExcelTools"
src = os.path.join(root, "build", "appicon.png")
ico = os.path.join(root, "build", "windows", "icon.ico")

img = Image.open(src).convert("RGBA")
sizes = [(16, 16), (24, 24), (32, 32), (48, 48), (64, 64), (128, 128), (256, 256)]
img.save(ico, format="ICO", sizes=sizes)
print("wrote", ico, "bytes", os.path.getsize(ico))
print("sizes", sizes)
