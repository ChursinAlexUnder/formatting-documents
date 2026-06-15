from docx.oxml.ns import qn
from docx.oxml import OxmlElement

def removeTabs(paragraph):
    """
    Удаляет все табуляции из указанного параграфа.
    """
    pPr = paragraph._element.find('.//w:pPr', namespaces={'w': 'http://schemas.openxmlformats.org/wordprocessingml/2006/main'})

    if pPr is not None:
        tabs = pPr.find('.//w:tabs', namespaces={'w': 'http://schemas.openxmlformats.org/wordprocessingml/2006/main'})

        if tabs is not None:
            pPr.remove(tabs)

def addTab(paragraph, listtabulation):
    """
    Добавляет табуляцию в параграф, если она ещё не добавлена.
    """
    tab_pos = 142 * int(float(listtabulation) / 0.25) - int(float(listtabulation) / 1)
    pPr = paragraph._element.find('.//w:pPr', namespaces={'w': 'http://schemas.openxmlformats.org/wordprocessingml/2006/main'})

    if pPr is not None:
        tabs = pPr.find('.//w:tabs', namespaces={'w': 'http://schemas.openxmlformats.org/wordprocessingml/2006/main'})

        if tabs is None:
            tabs = OxmlElement('w:tabs')
            pPr.append(tabs)
        tab = OxmlElement('w:tab')
        tab.set(qn('w:val'), 'left')
        tab.set(qn('w:pos'), str(tab_pos))
        tabs.append(tab)


def format_list_markers(document, font, fontsize):
    numbering_xml = document.part.numbering_part.element

    for abstract_num in numbering_xml.findall(qn("w:abstractNum")):
        for level in abstract_num.findall(qn("w:lvl")):
            number_format = level.find(qn("w:numFmt"))
            if number_format is None:
                continue

            run_properties = level.find(qn("w:rPr"))
            if run_properties is None:
                run_properties = OxmlElement("w:rPr")
                level.append(run_properties)

            if number_format.get(qn("w:val")) != "bullet":
                fonts = run_properties.find(qn("w:rFonts"))
                if fonts is None:
                    fonts = OxmlElement("w:rFonts")
                    run_properties.append(fonts)
                fonts.set(qn("w:ascii"), font)
                fonts.set(qn("w:hAnsi"), font)

            size = run_properties.find(qn("w:sz"))
            if size is None:
                size = OxmlElement("w:sz")
                run_properties.append(size)
            size.set(qn("w:val"), str(int(fontsize) * 2))
