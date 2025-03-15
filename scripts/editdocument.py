import sys
import os
from docx import Document
from docx.shared import Pt, Cm, RGBColor
from docx.enum.text import WD_PARAGRAPH_ALIGNMENT
from docx.oxml.ns import qn
from docx.oxml import OxmlElement

from modules.tabs import removeTabs, addTab
from modules.headings import headingLevel, isHeading, getDefaultFontSize, cycle_removeEmptyLinesAndPageBreaks, addPageBreak, addEmptyParagraphBefore, addEmptyParagraphAfter, ensureHeadingStyle, changeNormalStyle

def updateParagraphDefaultFont(paragraph, font):
    """Изменяет шрифт в первом <w:rFonts> внутри <w:pPr>, если он существует."""
    
    # Ищем <w:pPr> внутри параграфа
    pPr = paragraph._element.find(qn("w:pPr"))
    if pPr is not None:
        # Ищем <w:rPr> внутри <w:pPr>
        rPr = pPr.find(qn("w:rPr"))
        if rPr is not None:
            # Ищем <w:rFonts> внутри <w:rPr>
            rFonts = rPr.find(qn("w:rFonts"))
            if rFonts is not None:
                # Меняем шрифт только если <w:rFonts> уже есть
                rFonts.set(qn("w:ascii"), font)
                rFonts.set(qn("w:hAnsi"), font)

def cleanParagraphText(paragraph):
    """Удаляет пробелы в начале и в конце текста параграфа, 
    сохраняя изображения, гиперссылки и другие элементы."""

    if not paragraph.runs:  # Если нет run'ов, ничего не делаем
        return

    # --- Удаляем пробелы в начале ---
    found_non_space = False  # Флаг, нашли ли непустой текст
    for run in paragraph.runs:
        if not found_non_space:
            if run.text.strip():  # Нашли первый run с текстом
                run.text = run.text.lstrip()  # Убираем пробелы слева
                found_non_space = True  # Дальше пробелы не трогаем
            elif run.text.isspace():  
                run.text = ""  # Полностью пробельные run'ы в начале удаляем

    # --- Удаляем пробелы в конце ---
    found_non_space = False
    for run in reversed(paragraph.runs):
        if not found_non_space:
            if run.text.strip():  # Нашли последний run с текстом
                run.text = run.text.rstrip()  # Убираем пробелы справа
                found_non_space = True  # Дальше пробелы не трогаем
            elif run.text.isspace():  
                run.text = ""  # Полностью пробельные run'ы в конце удаляем

def modifyList(doc, font, fontsize):
    """
    Изменяет стиль номеров или маркеров списка в документе.
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

def formatDocument(bufferPath, documentName, font, fontsize, alignment, spacing, beforespacing, afterspacing, firstindentation, listtabulation, content):
    # Открываем документ
    doc = Document(bufferPath + '/' + documentName)

    defaultFontsize = getDefaultFontSize(doc, fontsize)

    haveList = False

    # Настройка полей для основного раздела
    for section in doc.sections:
        section.left_margin = Cm(3)
        section.right_margin = Cm(1.5)
        section.top_margin = Cm(2)
        section.bottom_margin = Cm(2)

    # задание настроек для обычного стандартного стиля Normal (в основном, для создаваемых отступов и разрывов страниц)
    changeNormalStyle(doc, font, fontsize, alignment, spacing, beforespacing, afterspacing, firstindentation)

    # обработка всего документа (по всем paragraphs и всем runs)
    for index, paragraph in enumerate(doc.paragraphs):
        isHead = False
        isDraw = False

        cleanParagraphText(paragraph)  # Убираем пробелы перед проверкой
        level = headingLevel(paragraph.text)

        # Проверяем наличие элемента <w:drawing> или <w:pict>
        for run in paragraph.runs:
            drawing = run._element.find(qn("w:drawing"))
            pict = run._element.find(qn("w:pict"))
            if drawing is not None or pict is not None:
                isDraw = True
                break

        if paragraph._element.xpath(".//w:numPr"):
            removeTabs(paragraph)  # Удаляем табуляции перед добавлением новой
            addTab(paragraph, listtabulation)
            if not haveList:
                haveList = True
                modifyList(doc, font, int(fontsize))
        elif not isDraw and isHeading(paragraph, index, doc, defaultFontsize) and level != False:
            style_name = ensureHeadingStyle(doc, level, font, fontsize)
            paragraph.style = style_name
            isHead = True

            cycle_removeEmptyLinesAndPageBreaks(doc, defaultFontsize)

            if level == 1:
                addPageBreak(paragraph)
                addEmptyParagraphAfter(paragraph)
            else:
                if index > 0:
                    addEmptyParagraphBefore(paragraph)
                addEmptyParagraphAfter(paragraph)

        # Доступ к низкоуровневому XML-элементу параграфа
        p = paragraph._element

        # Удаляем все элементы w:spacing, которые могут содержать интервалы
        for spacing_elem in p.xpath('.//w:spacing'):
            # Удаляем сам элемент
            spacing_elem.getparent().remove(spacing_elem)

        # Выравнивание текста
        if isDraw == True:
            paragraph.alignment = WD_PARAGRAPH_ALIGNMENT.CENTER
        elif isHead == True or alignment == "По левому краю":
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
        paragraph.paragraph_format.space_before = Pt(float(fontsize) * float(beforespacing))
        
        # интервал после абзаца
        paragraph.paragraph_format.space_after = Pt(float(fontsize) * float(afterspacing))
        
        # сбрасываем отступ всего абзаца
        paragraph.paragraph_format.left_indent = 0
        paragraph.paragraph_format.right_indent = 0
        
        # отступ первой строки
        if isDraw == True:
            paragraph.paragraph_format.first_line_indent = Cm(0)
        else:
            paragraph.paragraph_format.first_line_indent = Cm(float(firstindentation))

        updateParagraphDefaultFont(paragraph, font)
        
        # для уверенности
        paragraph.style.font.name = font
        paragraph.style.font.size = Pt(float(fontsize))
        
        for run in paragraph.runs:
            # Шрифт
            run.font.name = font
            # Размер шрифта
            run.font.size = Pt(float(fontsize))
            # цвет текста
            run.font.color.rgb = RGBColor(0, 0, 0)  # Чёрный цвет (RGB: 0, 0, 0)
    
    if content == "Добавить/обновить":
        # создание/обновление содержания!!!
        print("Hello")

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
content = sys.argv[10]

if not os.path.exists(documentPath):
    print(f"Document not found: {documentPath}")
    sys.exit(1)

try:
    formattedDocumentPath = formatDocument(bufferPath, documentName, font, fontsize, alignment, spacing, beforespacing, afterspacing, firstindentation, listtabulation, content)
    print(f"Formatted document created: {formattedDocumentPath}")
except Exception as e:
    print(f"Error formatting document: {e}")
    sys.exit(1)