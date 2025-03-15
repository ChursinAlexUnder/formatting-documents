import re
from docx.oxml.ns import qn

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

    # Формируем словарь с нумерацией и значениями False
    bibliography_dict = {i + 1: False for i in range(bibliography_length)}
    
    return bibliography_dict

def has_figure_reference(doc, start_index, figure_number):
    """
    Ищет ссылку на рисунок в виде (рисунок N), (рисунок N-M) или (рисунок N, M, ...)
    начиная с параграфа перед картинкой и до начала документа.
    """
    pattern = re.compile(r"\(рисунок\s+([\d,\-\s]+)\)", re.IGNORECASE)  

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

            if figure_number in numbers:
                return True

    return False

def find_pictures(doc):
    """Находит все рисунки в документе и возвращает словарь с их количеством."""
    figure_dict = {}
    figure_count = 0

    for index, paragraph in enumerate(doc.paragraphs):
        for run in paragraph.runs:
            drawing = run._element.find(qn("w:drawing"))
            pict = run._element.find(qn("w:pict"))
            if drawing is not None or pict is not None:
                figure_count += 1
                has_ref = has_figure_reference(doc, index, figure_count)
                figure_dict[figure_count] = has_ref

    return figure_dict

# Аналогичную функцию сделать для таблиц!!!