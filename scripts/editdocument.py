import sys
import os
from docx import Document
from docx.shared import Pt
from docx.enum.text import WD_PARAGRAPH_ALIGNMENT
from docx.shared import Cm
from docx.oxml.ns import qn
from docx.oxml import OxmlElement

def modify_list_numbering_style(doc, font, fontsize):
    """
    Изменяет стиль номеров или маркеров списка в документе.
    Для ненумерованных списков (bullet) изменяется только размер шрифта.
    """
    # Доступ к part с нумерациями
    numbering_part = doc.part.numbering_part
    numbering_xml = numbering_part.element

    # Получаем все уровни списков (abstractNum -> lvl)
    for abstract_num in numbering_xml.findall(qn("w:abstractNum")):
        for lvl in abstract_num.findall(qn("w:lvl")):
            # Проверяем тип списка
            num_fmt = lvl.find(qn("w:numFmt"))
            if num_fmt is None:
                continue

            num_fmt_val = num_fmt.get(qn("w:val"))
            
            # Устанавливаем rPr для уровня списка
            rPr = lvl.find(qn("w:rPr"))
            if rPr is None:
                rPr = OxmlElement("w:rPr")
                lvl.append(rPr)

            # Изменяем стиль в зависимости от типа списка
            if num_fmt_val != "bullet":  # Нумерованный список
                # Устанавливаем шрифт
                rFonts = rPr.find(qn("w:rFonts"))
                if rFonts is None:
                    rFonts = OxmlElement("w:rFonts")
                    rPr.append(rFonts)
                rFonts.set(qn("w:ascii"), font)
                rFonts.set(qn("w:hAnsi"), font)

            # Устанавливаем размер шрифта
            sz = rPr.find(qn("w:sz"))
            if sz is None:
                sz = OxmlElement("w:sz")
                rPr.append(sz)
            sz.set(qn("w:val"), str(fontsize * 2))

def formatDocument(bufferPath, documentName, font, fontsize, alignment, spacing, beforespacing, afterspacing, firstindentation):
    # Открываем документ
    doc = Document(bufferPath + '/' + documentName)

    haveList = False

    modify_list_numbering_style(doc, font, int(fontsize))

    # обработка всего документа (по всем paragraphs и всем runs)
    for paragraph in doc.paragraphs:

        if paragraph._element.xpath(".//w:numPr") and not haveList:
            haveList = True
            modify_list_numbering_style(doc, font, int(fontsize))

        # Доступ к низкоуровневому XML-элементу параграфа
        p = paragraph._element

        # Удаляем все элементы w:spacing, которые могут содержать интервалы
        for spacing_elem in p.xpath('.//w:spacing'):
            # Удаляем сам элемент
            spacing_elem.getparent().remove(spacing_elem)

        # Выравнивание текста
        if alignment == "По левому краю":
            paragraph.alignment = WD_PARAGRAPH_ALIGNMENT.LEFT
        elif alignment == "По центру":
            paragraph.alignment = WD_PARAGRAPH_ALIGNMENT.CENTER
        elif alignment == "По правому краю":
            paragraph.alignment = WD_PARAGRAPH_ALIGNMENT.RIGHT
        elif alignment == "По ширине":
            paragraph.alignment = WD_PARAGRAPH_ALIGNMENT.JUSTIFY

        # Междустрочный интервал
        paragraph.paragraph_format.line_spacing = float(spacing)

        # интервал перед абзацем
        if beforespacing == "Нет":
            paragraph.paragraph_format.space_before = Pt(0)
        else:
            paragraph.paragraph_format.space_before = Pt(float(fontsize) * float(beforespacing))
        
        # интервал после абзаца
        if afterspacing == "Нет":
            paragraph.paragraph_format.space_after = Pt(0)
        else:
            paragraph.paragraph_format.space_after = Pt(float(fontsize) * float(afterspacing))
        
        # отступ первой строки
        paragraph.paragraph_format.first_line_indent = Cm(float(firstindentation))
        
        for run in paragraph.runs:
            # Шрифт
            run.font.name = font
            # Размер шрифта
            run.font.size = Pt(float(fontsize))

        # Если параграф часть списка
        # if paragraph.style.name.startswith('List'):
            # 
        


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