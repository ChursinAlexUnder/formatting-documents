import re
from docx.oxml.ns import qn
from docx.enum.text import WD_PARAGRAPH_ALIGNMENT
from docx.shared import Cm

ns = {'w': 'http://schemas.openxmlformats.org/wordprocessingml/2006/main'}

def findBibliographyList(doc):
    """Находит список литературы в конце документа и возвращает словарь с его длиной."""
    flag = False
    bibliographyStart = None
    bibliographyLength = 0
    paragraphs = doc.paragraphs[::-1]  # Проходим документ с конца

    for i, para in enumerate(paragraphs):
        text = para.text.strip().lower()

        # Проверяем, является ли абзац заголовком списка литературы
        if text.startswith("список") and ("источников" in text or "литературы" in text):
            bibliographyStart = len(doc.paragraphs) - 1 - i
            break

    if bibliographyStart is not None:
        # Теперь считаем абзацы списка литературы
        for i in range(bibliographyStart + 1, len(doc.paragraphs)):
            para = doc.paragraphs[i]
            text = para.text.strip()

            # Если встретили заголовок — останавливаемся
            if para.style.name.startswith("Heading") or para.style.name.startswith("Заголовок"):
                break

            for run in para.runs:
                brs = run._element.findall(".//w:br", namespaces=ns)
                for br in brs:
                    if br.get(qn("w:type")) == "page":
                        flag = True
                        break
                if flag == True:
                    break
            if flag == True:
                break

            if text:
                bibliographyLength += 1

    bibliographyList = [False] * bibliographyLength
    return bibliographyList

def hasReference(doc, start_index, number, pattern):
    """
    Ищет ссылку на рисунок в виде (рисунок N), (рисунок N-M) или (рисунок N, M, ...)
    начиная с параграфа перед картинкой и до начала документа.
    """
    for i in range(start_index - 1, -1, -1):
        matches = pattern.findall(doc.paragraphs[i].text)
        for match in matches:
            numbers = set()
            # Сначала разбиваем содержимое ссылки по запятым
            tokens = match.split(',')
            for token in tokens:
                token = token.strip()
                # Если токен содержит тире (обычный, en dash или em dash)
                if any(dash in token for dash in ["-", "–", "—"]):
                    try:
                        # Разбиваем токен по тире с учётом пробелов
                        dash_split = re.split(r"\s*[-–—]\s*", token)
                        if len(dash_split) == 2:
                            start, end = map(int, dash_split)
                            numbers.update(range(start, end + 1))
                    except ValueError:
                        continue
                else:
                    try:
                        numbers.add(int(token))
                    except ValueError:
                        continue
            if number in numbers:
                return True
    return False

def getTableParagraphIndex(doc, table):
    """
    Определяет индекс параграфа, непосредственно предшествующего таблице.
    """
    last_paragraph_index = None
    p_index = 0
    # Проходим по всем дочерним элементам тела документа
    for child in doc.element.body:
        tag = child.tag
        if tag.endswith('}p'):
            last_paragraph_index = p_index
            p_index += 1
        elif tag.endswith('}tbl'):
            if child == table._element:
                break
    if last_paragraph_index is None:
        last_paragraph_index = len(doc.paragraphs) - 1
    return last_paragraph_index

def findAndFormatTables(doc):
    """
    Находит все таблицы в документе, форматирует их вместе с заголовками и возвращает список булевых значений,
    соответствующих наличию ссылки на таблицу.
    """
    tableList = []
    tableCount = 0
    tablePattern = re.compile(r"(?:\(\s*)?таблиц\w*\s+([\d,\-–—\s]+?)(?:\s*\))?", re.IGNORECASE)

    for table in doc.tables:
        tableCount += 1
        tableParagraphIndex = getTableParagraphIndex(doc, table)
        # Если предыдущий параграф существует, выравниваем его по центру.
        
        if 0 <= tableParagraphIndex < len(doc.paragraphs) and doc.paragraphs[tableParagraphIndex].text.strip().lower().startswith("таблиц"):
            doc.paragraphs[tableParagraphIndex].paragraph_format.first_line_indent = Cm(0)
            doc.paragraphs[tableParagraphIndex].alignment = WD_PARAGRAPH_ALIGNMENT.LEFT
        hasRef = hasReference(doc, tableParagraphIndex, tableCount, tablePattern)
        tableList.append(hasRef)
    return tableList