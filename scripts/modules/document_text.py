import math
import re
import unicodedata

from docx.oxml import OxmlElement
from docx.oxml.ns import qn


XML_SPACE = "{http://www.w3.org/XML/1998/namespace}space"
FIGURE_CAPTION_RE = re.compile(
    r"^(\s*)(рис(?:унок|\.?|\s*-?(?:ке|ках|ку|нок))?|рисунки)(\s+\d+(?:\.\d+)*(?:\s*(?:,|-|–|—)\s*\d+(?:\.\d+)*)*)(?:(\s+[-–]\s+)(.*)|(\s+)(.*))?$",
    re.IGNORECASE,
)
TABLE_CAPTION_RE = re.compile(
    r"^(\s*)(таб(?:лица|лицы|л\.?|\.?|\s*-?(?:ца|цы|це|цу|л)))(\s+\d+(?:\.\d+)*(?:\s*(?:,|-|–|—)\s*\d+(?:\.\d+)*)*)(?:(\s+[-–]\s+)(.*)|(\s+)(.*))?$",
    re.IGNORECASE,
)
CONTINUATION_TABLE_HEADING_RE = re.compile(
    r"^(?=.*\b(?:продолжени(?:е|я)?)\b)(?=.*\b(?:таб(?:лица|лицы|л\.?|\.?|\s*-?(?:ца|цы|це|цу|л)))\b).*$",
    re.IGNORECASE,
)
NUMBERED_HEADING_RE = re.compile(
    r"^(\s*\d+(?:\.\d+){0,3}(?:[.)])?\s+)(.*)$",
    re.DOTALL,
)
LINE_WIDTH_SAFETY_FACTOR = 0.985
APPENDIX_HEADING_RE = re.compile(
    r"^\s*приложение(?:\s+[A-Za-zА-Яа-яЁё0-9]+)(?:\s|$)",
    re.IGNORECASE,
)


def replace_em_dashes(document):
    """Заменяет длинное тире во всех доступных текстовых XML-частях DOCX."""
    for part in document.part.package.parts:
        element = getattr(part, "element", None)
        if element is None:
            continue
        for text_node in element.iter(qn("w:t")):
            if text_node.text and "—" in text_node.text:
                text_node.text = text_node.text.replace("—", "–")


def normalize_heading(paragraph):
    """Нормализует регистр после номера и удаляет пунктуацию в конце."""
    text = _paragraph_text(paragraph)
    match = NUMBERED_HEADING_RE.match(text)
    if match:
        text = match.group(1) + _uppercase_first_letter(match.group(2))
    text = _strip_trailing_punctuation(text)
    _replace_paragraph_text(paragraph, text)


def normalize_caption(paragraph):
    """Нормализует подпись рисунка/таблицы, тире, регистр и окончание."""
    text = _paragraph_text(paragraph)
    if is_table_continuation_heading(text):
        normalized = re.sub(r"\s+", " ", text).strip()
        number_match = re.search(r"(?:\(|\b)(\d+(?:\.\d+)*)(?:\)|\b)?$", normalized)
        number = number_match.group(1) if number_match else ""
        normalized = f"Продолжение таблицы {number}" if number else "Продолжение таблицы"
        if normalized != text:
            _replace_paragraph_text(paragraph, normalized)
        return

    figure_match = FIGURE_CAPTION_RE.match(text)
    table_match = TABLE_CAPTION_RE.match(text)
    match = figure_match or table_match
    if not match:
        return

    leading, label, numbers = match.group(1), match.group(2), match.group(3)
    separator = match.group(4)
    caption = match.group(5)
    whitespace = match.group(6)
    text_without_separator = match.group(7)

    if re.search(r"\bрис(?:унок|\.?|\s*-?ке)?\b", label, re.IGNORECASE):
        canonical_label = "Рисунок"
        plural_label = "Рисунки"
    else:
        canonical_label = "Таблица"
        plural_label = "Таблицы"

    normalized_numbers = _normalize_number_range(numbers)
    if re.search(r"\b(?:рисунки|таблицы)\b", label, re.IGNORECASE):
        label = plural_label
    else:
        label = canonical_label

    if separator is not None:
        normalized = f"{leading}{label} {normalized_numbers} – {_uppercase_first_letter(caption or '')}"
    elif text_without_separator is not None:
        normalized = (
            f"{leading}{label} {normalized_numbers} – "
            f"{_uppercase_first_letter(text_without_separator)}"
        )
    else:
        normalized = f"{leading}{label} {normalized_numbers}"

    normalized = _strip_trailing_punctuation(normalized)
    _replace_paragraph_text(paragraph, normalized)


def is_figure_caption(text):
    return bool(FIGURE_CAPTION_RE.match(text or ""))


def is_table_caption(text):
    return bool(TABLE_CAPTION_RE.match(text or ""))


def is_table_continuation_heading(text):
    if not text:
        return False
    normalized_text = " ".join(text.split())
    return bool(CONTINUATION_TABLE_HEADING_RE.match(normalized_text))


def _normalize_number_range(text):
    if not text:
        return ""
    normalized = re.sub(r"\s+", "", text)
    normalized = normalized.replace("—", "-").replace("–", "-")
    parts = []
    for part in normalized.split(","):
        part = part.strip()
        if not part:
            continue
        if "-" in part:
            parts.append(part)
        else:
            parts.append(part)
    if len(parts) > 1:
        return ", ".join(parts)
    return "".join(parts)


def normalize_parenthetical_references(text):
    """Нормализует сокращённые ссылки в скобках к полному виду."""
    if not text:
        return ""

    def replace_match(match):
        kind = match.group(1).strip().lower()
        numbers = match.group(2).strip()
        normalized_numbers = _normalize_number_range(numbers)
        if kind.startswith("рис"):
            label = "рисунок" if "," not in normalized_numbers and "-" not in normalized_numbers else "рисунки"
            return f"({label} {normalized_numbers})"
        label = "таблица" if "," not in normalized_numbers and "-" not in normalized_numbers else "таблицы"
        return f"({label} {normalized_numbers})"

    return re.sub(
        r"\(((?:рис|таб)[^\s)]*)\s+"
        r"(\d+(?:\.\d+)*(?:\s*(?:,|-|–|—)\s*\d+(?:\.\d+)*)*)\)",
        replace_match,
        text,
        flags=re.IGNORECASE,
    )


def normalize_paragraph_references(paragraph):
    original = _paragraph_text(paragraph)
    normalized = normalize_parenthetical_references(original)
    if normalized != original:
        _replace_paragraph_text(paragraph, normalized)


def is_appendix_heading(text):
    return bool(APPENDIX_HEADING_RE.match(text or ""))


def normalize_appendix_heading(paragraph):
    text = _paragraph_text(paragraph).strip().upper()
    _replace_paragraph_text(paragraph, text)


def trim_paragraph_text(paragraph):
    if not paragraph.runs:
        return

    found_text = False
    for run in paragraph.runs:
        if found_text:
            break
        if run.text.strip():
            run.text = run.text.lstrip()
            found_text = True
        elif run.text.isspace():
            run.text = ""

    found_text = False
    for run in reversed(paragraph.runs):
        if found_text:
            break
        if run.text.strip():
            run.text = run.text.rstrip()
            found_text = True
        elif run.text.isspace():
            run.text = ""


def ensure_paragraph_mark_formatting(p_element, font, fontsize):
    """Задаёт шрифт знаку абзаца, включая полностью пустые строки."""
    p_pr = p_element.find(qn("w:pPr"))
    if p_pr is None:
        p_pr = OxmlElement("w:pPr")
        p_element.insert(0, p_pr)

    r_pr = p_pr.find(qn("w:rPr"))
    if r_pr is None:
        r_pr = OxmlElement("w:rPr")
        p_pr.append(r_pr)

    r_fonts = r_pr.find(qn("w:rFonts"))
    if r_fonts is None:
        r_fonts = OxmlElement("w:rFonts")
        r_pr.append(r_fonts)
    for attribute in ("ascii", "hAnsi", "eastAsia", "cs"):
        r_fonts.set(qn(f"w:{attribute}"), font)

    half_points = str(int(round(float(fontsize) * 2)))
    for tag in ("w:sz", "w:szCs"):
        size = r_pr.find(qn(tag))
        if size is None:
            size = OxmlElement(tag)
            r_pr.append(size)
        size.set(qn("w:val"), half_points)


def set_paragraph_spacing(paragraph, before_points, after_points, line_spacing):
    """Жёстко переопределяет интервалы, включая автозначения из web-стилей."""
    p_element = getattr(paragraph, "_element", paragraph)
    p_pr = p_element.find(qn("w:pPr"))
    if p_pr is None:
        p_pr = OxmlElement("w:pPr")
        p_element.insert(0, p_pr)

    spacing = p_pr.find(qn("w:spacing"))
    if spacing is None:
        spacing = OxmlElement("w:spacing")
        p_pr.append(spacing)

    for attribute in (
        "before",
        "beforeLines",
        "beforeAutospacing",
        "after",
        "afterLines",
        "afterAutospacing",
        "line",
        "lineRule",
    ):
        qualified = qn(f"w:{attribute}")
        if qualified in spacing.attrib:
            del spacing.attrib[qualified]

    spacing.set(qn("w:before"), str(int(round(float(before_points) * 20))))
    spacing.set(qn("w:after"), str(int(round(float(after_points) * 20))))
    spacing.set(qn("w:beforeAutospacing"), "0")
    spacing.set(qn("w:afterAutospacing"), "0")
    spacing.set(qn("w:line"), str(int(round(float(line_spacing) * 240))))
    spacing.set(qn("w:lineRule"), "auto")

    contextual_spacing = p_pr.find(qn("w:contextualSpacing"))
    if contextual_spacing is not None:
        p_pr.remove(contextual_spacing)


def paragraph_will_wrap(document, paragraph, font, fontsize, bold=False):
    """
    Оценивает перенос Word по полезной ширине страницы и ширине символов.
    Точный результат зависит от установленного в Word шрифта, поэтому
    используется небольшой запас по ширине.
    """
    text = paragraph.text.strip()
    if not text:
        return False

    if "\n" in text or paragraph._element.xpath(
        ".//w:br[not(@w:type='page')]"
    ):
        return True

    section = _paragraph_section(document, paragraph)
    available_width = (
        section.page_width.pt
        - section.left_margin.pt
        - section.right_margin.pt
    )

    paragraph_format = paragraph.paragraph_format
    for indent in (paragraph_format.left_indent, paragraph_format.right_indent):
        if indent is not None:
            available_width -= max(indent.pt, 0)

    first_line_width = available_width
    if paragraph_format.first_line_indent is not None:
        first_line_width -= max(paragraph_format.first_line_indent.pt, 0)
    available_width *= LINE_WIDTH_SAFETY_FACTOR
    first_line_width *= LINE_WIDTH_SAFETY_FACTOR
    return _estimate_line_count(
        text,
        available_width,
        float(fontsize),
        font,
        bold,
        first_line_width=first_line_width,
    ) > 1


def _paragraph_section(document, paragraph):
    section_index = 0
    for child in document.element.body.iterchildren():
        if child is paragraph._element:
            break
        if child.tag == qn("w:p"):
            section_properties = child.find("./w:pPr/w:sectPr", child.nsmap)
            if section_properties is not None:
                section_index += 1

    return document.sections[min(section_index, len(document.sections) - 1)]


def _estimate_line_count(
    text,
    available_width,
    fontsize,
    font,
    bold,
    first_line_width=None,
):
    if available_width <= 0:
        return 2

    if first_line_width is None:
        first_line_width = available_width
    if first_line_width <= 0:
        return 2

    lines = 1
    current_width = 0.0
    tokens = re.findall(r"\S+\s*", text)

    for token in tokens:
        token_width = _text_width(token, fontsize, font, bold)
        line_width = first_line_width if lines == 1 else available_width
        if current_width and current_width + token_width > line_width:
            lines += 1
            current_width = token_width
        else:
            current_width += token_width

        line_width = first_line_width if lines == 1 else available_width
        if token_width > line_width:
            extra_lines = max(0, math.ceil(token_width / line_width) - 1)
            lines += extra_lines
            current_width = token_width % available_width

    return lines


def _text_width(text, fontsize, font, bold):
    font_scale = {
        "arial": 0.98,
        "calibri": 0.96,
        "courier new": 1.12,
        "georgia": 1.03,
        "tahoma": 0.97,
        "times new roman": 1.0,
        "verdana": 1.08,
    }.get(font.lower(), 1.0)
    bold_scale = 1.04 if bold else 1.0
    return sum(_character_width_factor(char) for char in text) * fontsize * font_scale * bold_scale


def _character_width_factor(char):
    if char.isspace():
        return 0.28
    if char in ".,:;!|iIl1'`":
        return 0.27
    if char in "mwMWЖШЩЮФжшщюф":
        return 0.82
    if char.isupper():
        return 0.64
    if char.isdigit():
        return 0.52
    if unicodedata.category(char).startswith("P"):
        return 0.34
    return 0.52


def _strip_trailing_punctuation(text):
    text = text.rstrip()
    while text and unicodedata.category(text[-1]).startswith("P"):
        text = text[:-1].rstrip()
    return text


def _uppercase_first_letter(text):
    chars = list(text)
    for index, char in enumerate(chars):
        if char.isalpha():
            chars[index] = char.upper()
            break
    return "".join(chars)


def _paragraph_text(paragraph):
    return "".join(node.text or "" for node in paragraph._element.iter(qn("w:t")))


def _replace_paragraph_text(paragraph, new_text):
    text_nodes = list(paragraph._element.iter(qn("w:t")))
    if not text_nodes:
        return

    remaining = new_text
    for index, node in enumerate(text_nodes):
        if index == len(text_nodes) - 1:
            value = remaining
        else:
            original_length = len(node.text or "")
            value = remaining[:original_length]
            remaining = remaining[original_length:]

        node.text = value
        if value.startswith(" ") or value.endswith(" "):
            node.set(XML_SPACE, "preserve")
        elif XML_SPACE in node.attrib:
            del node.attrib[XML_SPACE]
