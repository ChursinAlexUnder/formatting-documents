import sys
import os
from docx import Document
from docx.shared import Pt
from docx.enum.text import WD_PARAGRAPH_ALIGNMENT
from docx.shared import Cm
from docx.oxml.ns import qn
from docx.oxml import OxmlElement

def remove_tabs_from_paragraph(paragraph):
    """
    Удаляет все табуляции из указанного параграфа.
    """
    # Получаем элемент w:pPr (настройки параграфа)
    pPr = paragraph._element.find('.//w:pPr', namespaces={'w': 'http://schemas.openxmlformats.org/wordprocessingml/2006/main'})
    
    if pPr is not None:
        # Ищем элемент w:tabs
        tabs = pPr.find('.//w:tabs', namespaces={'w': 'http://schemas.openxmlformats.org/wordprocessingml/2006/main'})
        
        if tabs is not None:
            # Удаляем найденный элемент w:tabs
            pPr.remove(tabs)

def add_tab_in_paragraph(paragraph, listtabulation):
    """
    Добавляет табуляцию в параграф, если она ещё не добавлена.
    """
    tab_pos = 142 * int(float(listtabulation) / 0.25) - int(float(listtabulation) / 1)
    
    # Получаем элемент w:pPr (параграф)
    pPr = paragraph._element.find('.//w:pPr', namespaces={'w': 'http://schemas.openxmlformats.org/wordprocessingml/2006/main'})
    
    if pPr is not None:
        # Проверяем, есть ли уже tabs
        tabs = pPr.find('.//w:tabs', namespaces={'w': 'http://schemas.openxmlformats.org/wordprocessingml/2006/main'})
        
        if tabs is None:
            # Если нет, создаем новый элемент w:tabs
            tabs = OxmlElement('w:tabs')
            pPr.append(tabs)
        
        # Добавляем новую табуляцию
        tab = OxmlElement('w:tab')
        tab.set(qn('w:val'), 'left')
        tab.set(qn('w:pos'), str(tab_pos))
        tabs.append(tab)

def modify_list_numbering_style(doc, font, fontsize):
    """
    Изменяет стиль номеров или маркеров списка в документе.
    Также изменяет табуляцию (расстояние между маркером и текстом списка).
    """
    numbering_part = doc.part.numbering_part
    numbering_xml = numbering_part.element

    for abstract_num in numbering_xml.findall(qn("w:abstractNum")):
        for lvl in abstract_num.findall(qn("w:lvl")):
            num_fmt = lvl.find(qn("w:numFmt"))
            if num_fmt is None:
                continue

            num_fmt_val = num_fmt.get(qn("w:val"))

            # Настройки шрифта
            rPr = lvl.find(qn("w:rPr"))
            if rPr is None:
                rPr = OxmlElement("w:rPr")
                lvl.append(rPr)

            if num_fmt_val != "bullet": #нумерованный список
                rFonts = rPr.find(qn("w:rFonts"))
                if rFonts is None:
                    rFonts = OxmlElement("w:rFonts")
                    rPr.append(rFonts)
                rFonts.set(qn("w:ascii"), font)
                rFonts.set(qn("w:hAnsi"), font)

            sz = rPr.find(qn("w:sz"))
            if sz is None:
                sz = OxmlElement("w:sz")
                rPr.append(sz)
            sz.set(qn("w:val"), str(fontsize * 2))

def formatDocument(bufferPath, documentName, font, fontsize, alignment, spacing, beforespacing, afterspacing, firstindentation, listtabulation):
    # Открываем документ
    doc = Document(bufferPath + '/' + documentName)

    haveList = False

    # Настройка полей для основного раздела
    for section in doc.sections:
        section.left_margin = Cm(3)
        section.right_margin = Cm(1.5)
        section.top_margin = Cm(2)
        section.bottom_margin = Cm(2)

    # обработка всего документа (по всем paragraphs и всем runs)
    for paragraph in doc.paragraphs:

        if paragraph._element.xpath(".//w:numPr"):
            remove_tabs_from_paragraph(paragraph)  # Удаляем табуляции перед добавлением новой
            add_tab_in_paragraph(paragraph, listtabulation)
            if not haveList:
                haveList = True
                # delete_tabulation(doc)
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
        
        # сбрасываем отступ всего абзаца
        paragraph.paragraph_format.left_indent = 0
        paragraph.paragraph_format.right_indent = 0
        
        # отступ первой строки
        paragraph.paragraph_format.first_line_indent = Cm(float(firstindentation))

        for run in paragraph.runs:
            # Шрифт
            run.font.name = font
            # Размер шрифта
            run.font.size = Pt(float(fontsize))
        
        # Проверяем наличие элемента <w:drawing> или <w:pict>
        for run in paragraph.runs:
            drawing = run._element.find(qn("w:drawing"))
            pict = run._element.find(qn("w:pict"))
            if drawing is not None or pict is not None:
                # Устанавливаем выравнивание параграфа по центру
                paragraph.alignment = WD_PARAGRAPH_ALIGNMENT.CENTER
                # отступ первой строки для картинки
                paragraph.paragraph_format.first_line_indent = Cm(0)
                break


    # Работа с именем отформатированного документа
    formattedDocumentName = 'formatted_' + documentName
    formattedDocumentPath = bufferPath + '/' + formattedDocumentName
    doc.save(formattedDocumentPath)
    
    return formattedDocumentPath

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
listtabulation = sys.argv[9]

if not os.path.exists(documentPath):
    print(f"Document not found: {documentPath}")
    sys.exit(1)

try:
    formattedDocumentPath = formatDocument(bufferPath, documentName, font, fontsize, alignment, spacing, beforespacing, afterspacing, firstindentation, listtabulation)
    print(f"Formatted document created: {formattedDocumentPath}")
except Exception as e:
    print(f"Error formatting document: {e}")
    sys.exit(1)