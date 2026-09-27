#!/usr/bin/env python3
"""Наполняет продовую БД демо-данными (СПб-компании, вакансии, резюме).

Запуск:  BOT_TOKEN=<токен бота> python3 scripts/seed_demo.py
Идемпотентность: пользователи апсертятся, компании/вакансии создаются
только если у рекрутера ещё нет компаний, резюме — upsert.
Демо-пользователи имеют ID 900_000_001+ — легко отличить от реальных.
"""

import hashlib
import hmac
import json
import os
import sys
import time
import urllib.error
import urllib.parse
import urllib.request

BOT_TOKEN = os.environ.get("BOT_TOKEN")
BASE = os.environ.get("BASE_URL", "https://eclipse-sim.ru") + "/api/v1"

if not BOT_TOKEN:
    sys.exit("BOT_TOKEN env is required")


def sign_init_data(user: dict) -> str:
    params = {
        "user": json.dumps(user, ensure_ascii=False, separators=(",", ":")),
        "auth_date": str(int(time.time())),
        "query_id": "seed",
    }
    launch_params = "\n".join(f"{k}={v}" for k, v in sorted(params.items()))
    secret_key = hmac.new(b"WebAppData", BOT_TOKEN.encode(), hashlib.sha256).digest()
    signature = hmac.new(secret_key, launch_params.encode(), hashlib.sha256).hexdigest()
    raw = "&".join(f"{k}={urllib.parse.quote(v, safe='')}" for k, v in params.items())
    return raw + "&hash=" + signature


def call(method: str, path: str, token: str = None, payload: dict = None):
    data = json.dumps(payload).encode() if payload is not None else None
    req = urllib.request.Request(BASE + path, data=data, method=method,
                                 headers={"Content-Type": "application/json"})
    if token:
        req.add_header("Authorization", "Bearer " + token)
    try:
        return json.load(urllib.request.urlopen(req, timeout=15))
    except urllib.error.HTTPError as e:
        print(f"  !! {method} {path} -> {e.code}: {e.read().decode()}")
        raise


def login(user_id: int, first_name: str) -> str:
    init_data = sign_init_data({"id": user_id, "first_name": first_name, "language_code": "ru"})
    return call("POST", "/auth", payload={"initData": init_data})["token"]


RECRUITERS = [
    {
        "id": 900000001, "name": "Ольга", "company": {
            "name": "Нева Тех",
            "description": "Продуктовая ИТ-команда: fintech и логистика. Санкт-Петербург, 60 человек.",
            "website": "https://neva-tech.example.ru",
            "address": "Санкт-Петербург, ул. Рубинштейна, 3",
            "position": "hr",
        },
        "vacancies": [
            {"title": "Backend-разработчик (Go)", "required_skills": "go, postgres, kafka, grpc",
             "min_experience_months": 24, "city": "Санкт-Петербург", "work_format": "hybrid",
             "employment_type": "full_time", "salary_min": 180000, "salary_max": 280000,
             "response_ttl_hours": 48,
             "description": "Платёжное ядро: высокие нагрузки, своя экосистема микросервисов."},
            {"title": "Frontend-разработчик (React)", "required_skills": "react, typescript, vite",
             "min_experience_months": 12, "city": "Санкт-Петербург", "work_format": "remote",
             "employment_type": "full_time", "salary_min": 150000, "salary_max": 240000,
             "response_ttl_hours": 24,
             "description": "Кабинет клиента: графики, realtime, дизайн-система."},
            {"title": "Стажёр-аналитик", "required_skills": "sql, аналитика, excel",
             "min_experience_months": 0, "city": "Санкт-Петербург", "work_format": "onsite",
             "employment_type": "internship", "salary_min": 40000, "salary_max": 60000,
             "response_ttl_hours": 168,
             "description": "Стажировка 6 месяцев с шансом остаться в продуктовой команде."},
        ],
    },
    {
        "id": 900000002, "name": "Сергей", "company": {
            "name": "Северный Код",
            "description": "Студия заказной разработки: инфраструктура и автоматизация.",
            "website": "https://sevkod.example.ru",
            "address": "Санкт-Петербург, Ленина, 15",
            "position": "owner",
        },
        "vacancies": [
            {"title": "DevOps-инженер", "required_skills": "kubernetes, terraform, ansible, ci/cd",
             "min_experience_months": 36, "city": "Санкт-Петербург", "work_format": "hybrid",
             "employment_type": "contract", "salary_min": 220000, "salary_max": 300000,
             "response_ttl_hours": 72,
             "description": "Обслуживание кластеров клиентов, 5 проектов в работе."},
            {"title": "QA-инженер", "required_skills": "playwright, postman, sql",
             "min_experience_months": 12, "city": "Санкт-Петербург", "work_format": "onsite",
             "employment_type": "part_time", "salary_min": 90000, "salary_max": 130000,
             "response_ttl_hours": 48,
             "description": "Автоматизация e2e-тестов веб-приложений."},
        ],
    },
]

CANDIDATES = [
    {"id": 900000101, "first_name": "Иван", "last_name": "Петров", "username": "ivan_petrov_spb",
     "resume": {"title": "Backend-разработчик (Go)", "skills": "go, postgres, kafka, grpc, docker",
                "experience_months": 36, "about": "4 года пишу микросервисы, люблю простые решения.",
                "education": "СПбГУ, математико-механический факультет",
                "links": [{"type": "github", "url": "https://github.com/ivan-petrov"}],
                "city": "Санкт-Петербург", "work_format": "remote", "employment_type": "full_time",
                "salary_min": 200000, "salary_max": 300000}},
    {"id": 900000102, "first_name": "Мария", "last_name": "Смирнова", "username": "maria_smr",
     "resume": {"title": "Frontend-разработчик (React)", "skills": "react, typescript, vite, redux",
                "experience_months": 24, "about": "Делаю интерфейсы, которые не стыдно показать.",
                "education": "СПбПУ, программная инженерия",
                "links": [{"type": "github", "url": "https://github.com/maria-smr"},
                          {"type": "linkedin", "url": "https://linkedin.com/in/maria-smr"}],
                "city": "Санкт-Петербург", "work_format": "hybrid", "employment_type": "full_time",
                "salary_min": 150000, "salary_max": 220000}},
    {"id": 900000103, "first_name": "Алексей", "last_name": "Кузнецов", "username": "kuznec_qa",
     "resume": {"title": "QA-инженер", "skills": "playwright, postman, sql, python",
                "experience_months": 18, "about": "Автоматизирую то, что вручную делают третий раз.",
                "education": "ИТМО, информационные системы",
                "links": [{"type": "github", "url": "https://github.com/kuznec-qa"}],
                "city": "Санкт-Петербург", "work_format": "onsite", "employment_type": "full_time",
                "salary_min": 100000, "salary_max": 140000}},
    {"id": 900000104, "first_name": "Ольга", "last_name": "Иванова", "username": "olga_ivanova",
     "resume": {"title": "Системный аналитик", "skills": "sql, bpmn, rest, аналитика",
                "experience_months": 6, "about": "Перешла из поддержки в аналитику, быстро учусь.",
                "education": "СПбГЭУ, экономика и управление",
                "links": [],
                "city": "Санкт-Петербург", "work_format": "remote", "employment_type": "full_time",
                "salary_min": 80000, "salary_max": 120000}},
    {"id": 900000105, "first_name": "Дмитрий", "last_name": "Соколов", "username": "dsokolov_devops",
     "resume": {"title": "DevOps-инженер", "skills": "kubernetes, terraform, ansible, prometheus",
                "experience_months": 48, "about": "Инфраструктура как код, никаких ручных деплоев.",
                "education": "СПбПУ, прикладная математика",
                "links": [{"type": "github", "url": "https://github.com/dsokolov"}],
                "city": "Санкт-Петербург", "work_format": "remote", "employment_type": "contract",
                "salary_min": 250000, "salary_max": 350000}},
    {"id": 900000106, "first_name": "Анна", "last_name": "Морозова", "username": "anna_moroz",
     "resume": {"title": "Стажёр-разработчик", "skills": "python, sql, git",
                "experience_months": 0, "about": "3 курс, ищу первую стажировку в ИТ.",
                "education": "ИТМО, магистратура, информатика",
                "links": [{"type": "github", "url": "https://github.com/anna-moroz"}],
                "city": "Санкт-Петербург", "work_format": "onsite", "employment_type": "internship",
                "salary_min": 50000, "salary_max": 70000}},
]


def main():
    print("== Рекрутеры, компании, вакансии ==")
    for rec in RECRUITERS:
        token = login(rec["id"], rec["name"])
        existing = call("GET", "/my/companies", token=token)
        if existing:
            print(f"  [{rec['id']}] {rec['company']['name']}: уже есть, пропускаю")
            continue
        company = call("POST", "/companies", token=token, payload=rec["company"])
        print(f"  [{rec['id']}] компания «{company['name']}» id={company['id']}")
        for v in rec["vacancies"]:
            vacancy = call("POST", f"/companies/{company['id']}/vacancies", token=token, payload=v)
            print(f"    + вакансия «{vacancy['title']}» id={vacancy['id']}")

    print("== Кандидаты, резюме ==")
    for cand in CANDIDATES:
        token = login(cand["id"], cand["first_name"])
        role_req = {"role": "candidate", "acceptPersonalData": True}
        call("POST", "/me/role", token=token, payload=role_req)
        resume = dict(cand["resume"])
        resume["user_id"] = cand["id"]
        saved = call("PUT", "/my/resume", token=token, payload=resume)
        print(f"  [{cand['id']}] {cand['first_name']} {cand['last_name']}: резюме id={saved['id']} «{saved['title']}»")

    print("Готово.")


if __name__ == "__main__":
    main()
