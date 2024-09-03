import sys
from docx import Document

def extract_text(docx_file):
    doc = Document(docx_file)
    full_text = []
    for paragraph in doc.paragraphs:
        full_text.append(paragraph.text)
    return '\n'.join(full_text)

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: edit-document.py <docx_file>")
        sys.exit(1)

    docx_file = sys.argv[1]
    print(extract_text(docx_file))