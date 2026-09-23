"""Exports preserve the reviewed evidence, script, and draft provenance."""
from copy import deepcopy
from io import BytesIO
import json
import re

from docx import Document
from pypdf import PdfReader
import pytest

from butterfly.exports import render_export


@pytest.fixture
def meeting():
    return {
        "id": "meeting-export-test",
        "title": "Протокол: Қазақша кездесу & <обсуждение>",
        "date": "2026-09-23",
        "mode": "demo",
        "status": "needs_review",
        "summary": "Обсудили запуск.\nҚұжатты дайындау & проверить <план>.",
        "transcript": [
            {"id": "s1", "start": 0, "end": 7, "speaker": "Данияр",
             "text": "Данияр подготовит отчёт до 25 сентября. Құжат дайын болады."},
            {"id": "s2", "start": 7, "end": 12, "speaker": "Анна",
             "text": "Нужно проверить <смету> & бюджет.\nОтветственный ещё не выбран."},
        ],
        "actions": [
            {"id": "a1", "text": "Подготовить отчёт", "owner": "Данияр",
             "deadline": "2026-09-25", "evidence": "Данияр подготовит отчёт до 25 сентября.",
             "segment_id": "s1", "status": "confirmed"},
            {"id": "a2", "text": "Проверить <смету> & бюджет", "owner": None,
             "deadline": None, "evidence": "Нужно проверить <смету> & бюджет.",
             "segment_id": "s2", "status": "pending"},
        ],
    }


def docx_text(blob):
    document = Document(BytesIO(blob))
    parts = [paragraph.text for paragraph in document.paragraphs]
    for table in document.tables:
        parts.extend(cell.text for row in table.rows for cell in row.cells)
    return "\n".join(parts)


def pdf_text(blob):
    return "\n".join(page.extract_text() for page in PdfReader(BytesIO(blob)).pages)


def normalized(text):
    return re.sub(r"\s+", " ", text)


@pytest.mark.parametrize("format", ["pdf", "docx"])
def test_document_keeps_cyrillic_evidence_and_draft_provenance(meeting, format):
    before = deepcopy(meeting)
    blob = render_export(meeting, format)
    assert isinstance(blob, bytes)
    assert blob.startswith(b"%PDF" if format == "pdf" else b"PK")
    text = normalized(pdf_text(blob) if format == "pdf" else docx_text(blob))
    for expected in [
        "Қазақша кездесу", "<обсуждение>", "Құжатты дайындау",
        "Подготовить отчёт", "Данияр", "2026-09-25",
        "Данияр подготовит отчёт до 25 сентября.",
        "Проверить <смету> & бюджет", "Ответственный ещё не выбран.",
    ]:
        assert expected in text
    assert "ДЕМО" in text.upper() or "DEMO" in text.upper()
    assert "НЕ ПОДТВЕРЖДЕНО" in text.upper() or "UNCONFIRMED" in text.upper()
    assert meeting == before, "Export must not mutate the review state"


def test_json_export_round_trips_unknown_fields_and_multiline_text(meeting):
    result = json.loads(render_export(meeting, "json"))
    assert result["title"] == meeting["title"]
    assert result["summary"] == meeting["summary"]
    assert result["transcript"] == meeting["transcript"]
    assert result["actions"] == meeting["actions"]
    assert result["actions"][1]["owner"] is None
    assert result["actions"][1]["deadline"] is None


@pytest.mark.parametrize("format", ["pdf", "docx"])
def test_long_multiline_text_remains_readable_across_pages(meeting, format):
    meeting["transcript"] = [
        {"id": "long", "start": 0, "end": 300, "speaker": "Данияр",
         "text": "\n".join(f"Строка {n}: Құжат & <план> с доказательствами." for n in range(150))}
    ]
    blob = render_export(meeting, format)
    text = normalized(pdf_text(blob) if format == "pdf" else docx_text(blob))
    assert "Строка 0:" in text and "Строка 149:" in text
    if format == "pdf":
        assert len(PdfReader(BytesIO(blob)).pages) > 1


@pytest.mark.parametrize("format", ["pdf", "docx"])
def test_long_action_and_evidence_paginate_without_losing_the_last_line(meeting, format):
    action = meeting["actions"][0]
    action["text"] = "\n".join(f"Пункт {n}: проверить <план>." for n in range(90))
    action["evidence"] = "\n".join(f"Цитата {n}: Құжат & доказательство." for n in range(150))
    assert len(action["text"]) <= 4000 and len(action["evidence"]) <= 10000
    blob = render_export(meeting, format)
    text = normalized(pdf_text(blob) if format == "pdf" else docx_text(blob))
    assert "Пункт 89: проверить <план>." in text
    assert "Цитата 149: Құжат & доказательство." in text
    if format == "pdf":
        assert len(PdfReader(BytesIO(blob)).pages) > 1


@pytest.mark.parametrize("format", ["pdf", "docx", "json"])
def test_empty_transcript_exports_without_claiming_completion(meeting, format):
    meeting.update(mode="live", status="awaiting_provider", summary="", transcript=[], actions=[])
    blob = render_export(meeting, format)
    if format == "json":
        result = json.loads(blob)
        assert result["status"] == "awaiting_provider"
        assert result["transcript"] == []
    else:
        text = pdf_text(blob) if format == "pdf" else docx_text(blob)
        assert "Қазақша кездесу" in text
