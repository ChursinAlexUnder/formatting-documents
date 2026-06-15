import json
import os
import sys

from modules.annotation import build_annotation_source
from modules.document_formatter import format_document

formatDocument = format_document

def main():
    documentName = sys.argv[1]
    bufferPath = os.getenv('APP_BUFFER_DIR', '../buffer')
    documentPath = os.path.join(bufferPath, documentName)
    font = sys.argv[2]
    fontsize = sys.argv[3]
    alignment = sys.argv[4]
    spacing = sys.argv[5]
    beforespacing = sys.argv[6]
    afterspacing = sys.argv[7]
    firstindentation = sys.argv[8]
    listtabulation = sys.argv[9]
    havetitle = sys.argv[10]
    openrouter_api_key = (
        sys.argv[11]
        if len(sys.argv) > 11
        else os.getenv('OPENROUTER_API_KEY')
    )

    if not os.path.exists(documentPath):
        print(f"Документ не найден: {documentPath}", file=sys.stderr)
        sys.exit(1)

    try:
        formattedDocumentPath, result = format_document(
            bufferPath,
            documentName,
            font,
            fontsize,
            alignment,
            spacing,
            beforespacing,
            afterspacing,
            firstindentation,
            listtabulation,
            havetitle,
            openrouter_api_key,
        )
        print(json.dumps(result))
    except Exception as e:
        print(f"Ошибка форматирования документа: {e}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
