import sys
import os
from docx import Document
from docx.shared import Pt

def formatDocument(bufferPath, documentName, comment):
    # Открываем документ
    doc = Document(bufferPath + '/' + documentName)

    # обработка всего документа (по всем paragraphs и всем runs)
    for paragraph in doc.paragraphs:
        for run in paragraph.runs:
            run.font.size = Pt(14)
            run.font.name = 'Times New Roman'

    # Добавляем текст комментария в конец документа
    if comment:
        doc.add_paragraph(comment)
    else:
        doc.add_paragraph('Комментарий отсутствует, так что просто хорошего дня)')
    

    # Работа с именем отформатированного документа
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