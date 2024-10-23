import sys
import os
from docx import Document

def formatDocument(bufferPath, documentName, comment):
    # Открываем документ
    doc = Document(bufferPath + '/' + documentName)

    # Добавляем текст комментария в конец документа
    if comment:
        doc.add_paragraph(comment)
    else:
        doc.add_paragraph('Комментарий отсутствует, так что просто хорошего дня)')

    formattedDocumentName = 'formatted_' + documentName
    formattedDocumentPath = bufferPath + '/' + formattedDocumentName
    doc.save(formattedDocumentPath)
    
    return formattedDocumentPath

def main():
    documentName = sys.argv[1]
    bufferPath = '../buffer'
    documentPath = bufferPath + '/' + documentName
    comment = sys.argv[2]

    if not os.path.exists(documentPath):
        print(f"Document not found: {documentPath}")
        sys.exit(1)

    try:
        formattedDocumentPath = formatDocument(bufferPath, documentName, comment)
        print(f"Formatted document created: {formattedDocumentPath}")
    except Exception as e:
        print(f"Error formatting document: {e}")
        sys.exit(1)
main()