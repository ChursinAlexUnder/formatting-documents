import re
import time

import requests

from modules.document_text import is_figure_caption, is_table_caption
from modules.headings import isHeading
from modules.title import paragraphHasPageBreak
from modules.usednumbers import isBibliographyHeading


ANNOTATION_FALLBACK_MESSAGE = (
    "Аннотацию создать не удалось: сервис генерации временно недоступен. "
    "Документ отформатирован без аннотации."
)
ANNOTATION_SOURCE_LIMIT = 10000
ANNOTATION_FRAGMENT_LIMIT = 900
OPENROUTER_URL = "https://openrouter.ai/api/v1/chat/completions"
OPENROUTER_MODEL = "openrouter/owl-alpha"


def build_annotation_source(
    document,
    has_title_page=False,
    max_chars=ANNOTATION_SOURCE_LIMIT,
):
    paragraphs = document.paragraphs
    start_index = 0

    if has_title_page:
        for index, paragraph in enumerate(paragraphs[:40]):
            if paragraphHasPageBreak(paragraph):
                start_index = index + 1
                break

    candidates = []
    inside_contents = False

    for index in range(start_index, len(paragraphs)):
        paragraph = paragraphs[index]
        text = re.sub(r"\s+", " ", paragraph.text).strip()
        if not text:
            continue

        text_lower = text.lower()
        if isBibliographyHeading(text):
            break

        if text_lower in ("содержание", "оглавление"):
            inside_contents = True
            continue

        looks_like_toc_entry = bool(
            re.search(r"\.{2,}\s*\d+\s*$", text)
            or (
                inside_contents
                and re.search(r"\s+\d+\s*$", text)
                and len(text) < 180
            )
        )
        heading = _is_annotation_heading(paragraph, text)

        if inside_contents:
            if text_lower in ("аннотация", "реферат", "введение") or (
                heading and not looks_like_toc_entry
            ):
                inside_contents = False
            else:
                continue

        if looks_like_toc_entry:
            continue
        if is_figure_caption(text) or is_table_caption(text):
            continue
        if len(re.findall(r"[A-Za-zА-Яа-яЁё0-9]", text)) < 25 and not heading:
            continue

        candidates.append(
            {
                "index": index,
                "text": text,
                "lower": text_lower,
                "heading": heading,
            }
        )

    if not candidates:
        return ""

    selected = []
    selected_indices = set()

    def add_candidate(candidate):
        if candidate["index"] not in selected_indices:
            selected.append(candidate)
            selected_indices.add(candidate["index"])

    def add_section(keywords, body_count=3):
        for position, candidate in enumerate(candidates):
            if any(
                candidate["lower"] == keyword
                or candidate["lower"].startswith(keyword + " ")
                for keyword in keywords
            ):
                add_candidate(candidate)
                added_body = 0
                for following in candidates[position + 1:]:
                    if following["heading"] and added_body:
                        break
                    add_candidate(following)
                    if not following["heading"]:
                        added_body += 1
                    if added_body >= body_count:
                        break
                break

    add_section(("аннотация", "реферат"), 3)
    add_section(("введение",), 3)
    add_section(("заключение", "выводы"), 3)

    body_candidates = [
        candidate for candidate in candidates if not candidate["heading"]
    ]
    for candidate in body_candidates[:2]:
        add_candidate(candidate)
    for candidate in _distributed_items(body_candidates, 9):
        add_candidate(candidate)

    heading_candidates = [
        candidate for candidate in candidates if candidate["heading"]
    ]
    for candidate in _distributed_items(heading_candidates, 8):
        add_candidate(candidate)

    fragments = []
    current_length = 0
    document_span = max(len(paragraphs) - 1, 1)

    for candidate in selected:
        text = _truncate_fragment(
            candidate["text"],
            ANNOTATION_FRAGMENT_LIMIT,
        )
        position = round(candidate["index"] / document_span * 100)
        fragment = f"[Фрагмент, позиция около {position}%]\n{text}"
        separator_length = 2 if fragments else 0
        remaining = max_chars - current_length - separator_length
        if remaining <= 80:
            break
        if len(fragment) > remaining:
            fragment = _truncate_fragment(fragment, remaining)
        fragments.append(fragment)
        current_length += len(fragment) + separator_length

    return "\n\n".join(fragments)


def generate_annotation(document, api_key, has_title_page=False):
    if not api_key:
        return ""

    content_preview = build_annotation_source(document, has_title_page)
    if not content_preview.strip():
        return "Не удалось извлечь содержимое документа для генерации аннотации."

    prompt = (
        "Составь краткую аннотацию на русском языке по фрагментам документа. "
        "Фрагменты взяты из разных частей документа и снабжены примерной позицией. "
        "Определи общую тему, цель и содержание работы, не пересказывая структуру документа. "
        "Не акцентируй внимание на частных деталях, не используй перечисления через запятую. "
        "Аннотация должна быть одним абзацем, без заголовков, "
        "без пояснений и без списков. Длина — строго до 600 символов.\n\n"
        f"{content_preview}"
    )
    headers = {
        "Authorization": f"Bearer {api_key}",
        "Content-Type": "application/json",
        "HTTP-Referer": "https://formatting-documents.app",
        "X-Title": "Document Formatting Service",
    }
    data = {
        "model": OPENROUTER_MODEL,
        "messages": [
            {
                "role": "system",
                "content": (
                    "Ты пишешь краткие и точные аннотации к научным и "
                    "учебным текстам на русском языке."
                ),
            },
            {"role": "user", "content": prompt},
        ],
        "temperature": 0.2,
        "max_tokens": 300,
    }

    annotation = None
    for attempt in range(10):
        try:
            response = requests.post(
                OPENROUTER_URL,
                headers=headers,
                json=data,
                timeout=30,
            )
            if response.status_code == 429 or response.status_code >= 500:
                if attempt < 9:
                    time.sleep(3)
                    continue
                break
            response.raise_for_status()
            result = response.json()
            choices = result.get("choices", [])
            if choices:
                content = choices[0].get("message", {}).get("content", "")
                if isinstance(content, str) and content.strip():
                    annotation = content.strip()
            break
        except (requests.exceptions.RequestException, ValueError):
            if attempt < 9:
                time.sleep(3)

    if not annotation:
        return ANNOTATION_FALLBACK_MESSAGE

    if annotation.startswith("Аннотация"):
        annotation = annotation.split("\n", 1)[-1].strip()

    if len(annotation) > 600:
        truncated = annotation[:600]
        last_period = max(
            truncated.rfind("."),
            truncated.rfind("!"),
            truncated.rfind("?"),
        )
        annotation = (
            truncated[:last_period + 1]
            if last_period >= 0
            else truncated.rstrip()
        )

    return annotation


def _is_annotation_heading(paragraph, text):
    style_name = paragraph.style.name.lower() if paragraph.style else ""
    return (
        style_name.startswith("heading")
        or style_name.startswith("заголовок")
        or isHeading(paragraph)
        or text.lower()
        in (
            "аннотация",
            "реферат",
            "введение",
            "заключение",
            "выводы",
        )
    )


def _distributed_items(items, count):
    if not items or count <= 0:
        return []
    if len(items) <= count:
        return items
    if count == 1:
        return [items[len(items) // 2]]
    return [
        items[round(index * (len(items) - 1) / (count - 1))]
        for index in range(count)
    ]


def _truncate_fragment(text, limit):
    if len(text) <= limit:
        return text

    truncated = text[:limit].rstrip()
    sentence_end = max(
        truncated.rfind("."),
        truncated.rfind("!"),
        truncated.rfind("?"),
    )
    if sentence_end >= limit // 2:
        return truncated[:sentence_end + 1]

    last_space = truncated.rfind(" ")
    return truncated[:last_space].rstrip() if last_space > 0 else truncated
