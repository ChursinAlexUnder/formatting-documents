from docx.oxml.ns import qn

ns = {'w': 'http://schemas.openxmlformats.org/wordprocessingml/2006/main'}

def paragraphHasPageBreak(paragraph):
    """
    Функция проверяет, содержит ли параграф разрыв страницы.
    """
    for run in paragraph.runs:
        # Ищем элемент <w:br> с атрибутом w:type="page"
        brs = run._element.findall(".//w:br", namespaces=ns)
        for br in brs:
            if br.get(qn("w:type")) == "page":
                return True
    return False