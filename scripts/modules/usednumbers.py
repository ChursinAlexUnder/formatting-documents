from docx import Document
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

            # Если встретили разрыв страницы, заголовок или конец документа — останавливаемся
            if para.style.name.startswith("Heading"):
                break

            for run in para.runs:
                # Ищем элемент <w:br> с атрибутом w:type="page"
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