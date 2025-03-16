import re
from docx.oxml.ns import qn
from docx.enum.table import WD_TABLE_ALIGNMENT
from docx.enum.text import WD_PARAGRAPH_ALIGNMENT
from docx.shared import Cm

ns = {'w': 'http://schemas.openxmlformats.org/wordprocessingml/2006/main'}

def find_bibliography_list(doc):
    """Находит список литературы в конце документа и возвращает словарь с его длиной."""
    flag = False
    bibliography_start = None
    bibliography_length = 0
    paragraphs = doc.paragraphs[::-1]  # Проходим документ с конца

    for i, para in enumerate(paragraphs):
        text = para.text.strip().lower()

        # Проверяем, является ли абзац заголовком списка литературы
        if "список" in text and ("источников" in text or "литературы" in text):
            bibliography_start = len(doc.paragraphs) - 1 - i
            break

    if bibliography_start is not None:
        # Теперь считаем абзацы списка литературы
        for i in range(bibliography_start + 1, len(doc.paragraphs)):
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
                bibliography_length += 1

    # Формируем список из bibliography_length элементов, каждый из которых False
    bibliography_list = [False] * bibliography_length
    return bibliography_list

def has_reference(doc, start_index, number, pattern):
    """
    Ищет ссылку на рисунок в виде (рисунок N), (рисунок N-M) или (рисунок N, M, ...)
    начиная с параграфа перед картинкой и до начала документа.
    """
    for i in range(start_index - 1, -1, -1):  
        matches = pattern.findall(doc.paragraphs[i].text)  
        for match in matches:
            numbers = set()  
            parts = re.split(r"[, ]+", match.strip())

            for part in parts:
                if "-" in part:  
                    start, end = map(int, part.split("-"))  
                    numbers.update(range(start, end + 1))  
                else:
                    numbers.add(int(part))  

            if number in numbers:
                return True

    return False

def find_pictures(doc):
    """Находит все рисунки в документе и возвращает их массив"""
    figure_list = []
    figure_count = 0
    pattern = re.compile(r"(?:\(\s*)?рисун\w*\s+([\d,\-\s]+?)(?:\s*\))?", re.IGNORECASE)

    for index, paragraph in enumerate(doc.paragraphs):
        for run in paragraph.runs:
            drawing = run._element.find(qn("w:drawing"))
            pict = run._element.find(qn("w:pict"))
            if drawing is not None or pict is not None:
                figure_count += 1
                has_ref = has_reference(doc, index, figure_count, pattern)
                figure_list.append(has_ref)

    return figure_list

def get_table_paragraph_index(doc, table):
    """
    Определяет индекс параграфа, непосредственно предшествующего таблице.
    
    Для этого перебираем дочерние элементы тела документа (doc.element.body),
    считая параграфы. Когда встречается элемент таблицы, сравниваем его с
    table._element и возвращаем индекс последнего найденного параграфа.
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

def find_tables(doc):
    """
    Находит все таблицы в документе и возвращает список булевых значений,
    соответствующих наличию ссылки на таблицу.
    
    Для каждой таблицы определяется порядковый номер (по порядку появления)
    и индекс параграфа непосредственно перед таблицей, после чего вызывается
    has_reference для поиска ссылки.
    """
    table_list = []
    table_count = 0
    pattern = re.compile(r"(?:\(\s*)?таблиц\w*\s+([\d,\-\s]+?)(?:\s*\))?", re.IGNORECASE)

    for table in doc.tables:
        table_count += 1
        table_paragraph_index = get_table_paragraph_index(doc, table)
        has_ref = has_reference(doc, table_paragraph_index, table_count, pattern)
        table_list.append(has_ref)

    return table_list

def centerTableAndFormatTitle(doc):
    """
    Проходит по всем таблицам в документе и для каждой:
    - выравнивает таблицу по центру;
    - получает индекс параграфа, непосредственно предшествующего таблице,
      и форматирует его.
    """
    for table in doc.tables:
        # Получаем индекс параграфа, непосредственно предшествующего таблице.
        para_index = get_table_paragraph_index(doc, table)
        
        # Если предыдущий параграф существует, выравниваем его по центру.
        if 0 <= para_index < len(doc.paragraphs) and "таблиц" in doc.paragraphs[para_index].text.strip().lower():
            doc.paragraphs[para_index].paragraph_format.first_line_indent = Cm(0)
            doc.paragraphs[para_index].alignment = WD_PARAGRAPH_ALIGNMENT.LEFT