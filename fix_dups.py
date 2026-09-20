import re
filepath = "internal/rpc/api.go"
with open(filepath, "r", encoding="utf-8") as f:
    content = f.read()
    
# Remove duplicate cases
content = re.sub(r'case\s+"([^"]+)",\s*"\1":', r'case "\1":', content)

with open(filepath, "w", encoding="utf-8") as f:
    f.write(content)
