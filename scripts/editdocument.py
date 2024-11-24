import sys
import os
from docx import Document
from docx.shared import Pt
from docx.enum.text import WD_PARAGRAPH_ALIGNMENT
from docx.shared import Cm

def formatDocument(bufferPath, documentName, font, fontsize, alignment, spacing, beforespacing, afterspacing, firstindentation):
    # Открываем документ
    doc = Document(bufferPath + '/' + documentName)

    # обработка всего документа (по всем paragraphs и всем runs)
    for paragraph in doc.paragraphs:

        # Сохраняем текст параграфа в переменную
        paragraph_text = paragraph.text
        # Удаляем параграф
        paragraph._element.getparent().remove(paragraph._element)

        # Вставляем новый параграф с нужным текстом
        new_paragraph = doc.add_paragraph(paragraph_text)

        # Выравнивание текста
        if alignment == "По левому краю":
            new_paragraph.alignment = WD_PARAGRAPH_ALIGNMENT.LEFT
        elif alignment == "По центру":
            new_paragraph.alignment = WD_PARAGRAPH_ALIGNMENT.CENTER
        elif alignment == "По правому краю":
            new_paragraph.alignment = WD_PARAGRAPH_ALIGNMENT.RIGHT
        elif alignment == "По ширине":
            new_paragraph.alignment = WD_PARAGRAPH_ALIGNMENT.JUSTIFY

        # Междустрочный интервал
        new_paragraph.paragraph_format.line_spacing = float(spacing)

        # интервал перед абзацем
        if beforespacing == "Нет":
            new_paragraph.paragraph_format.space_before = Pt(0)
        else:
            new_paragraph.paragraph_format.space_before = Pt(float(fontsize) * float(beforespacing))
        
        # интервал после абзаца
        if afterspacing == "Нет":
            new_paragraph.paragraph_format.space_after = Pt(0)
        else:
            new_paragraph.paragraph_format.space_after = Pt(float(fontsize) * float(afterspacing))
        
        # отступ первой строки
        new_paragraph.paragraph_format.first_line_indent = Cm(float(firstindentation))
        
        for run in new_paragraph.runs:
            # Шрифт
            run.font.name = font

            # Размер шрифта
            run.font.size = Pt(float(fontsize))
            

    # Работа с именем отформатированного документа
    formattedDocumentName = 'formatted_' + documentName
    formattedDocumentPath = bufferPath + '/' + formattedDocumentName
    doc.save(formattedDocumentPath)
    
    return formattedDocumentPath

def main():
    documentName = sys.argv[1]
    bufferPath = '../buffer'
    documentPath = bufferPath + '/' + documentName
    font = sys.argv[2]
    fontsize = sys.argv[3]
    alignment = sys.argv[4]
    spacing = sys.argv[5]
    beforespacing = sys.argv[6]
    afterspacing = sys.argv[7]
    firstindentation = sys.argv[8]

    if not os.path.exists(documentPath):
        print(f"Document not found: {documentPath}")
        sys.exit(1)

    try:
        formattedDocumentPath = formatDocument(bufferPath, documentName, font, fontsize, alignment, spacing, beforespacing, afterspacing, firstindentation)
        print(f"Formatted document created: {formattedDocumentPath}")
    except Exception as e:
        print(f"Error formatting document: {e}")
        sys.exit(1)
main()