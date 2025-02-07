from docx.oxml.ns import qn
from docx.oxml import OxmlElement

def removeTabs(paragraph):
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

def addTab(paragraph, listtabulation):
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