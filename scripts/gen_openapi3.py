#!/usr/bin/env python3
"""Генерация openapi.yaml (OpenAPI 3.1) из живого swagger doc.json (Swagger 2.0).

Запуск: python3 scripts/gen_openapi3.py [url_or_file] [out]
По умолчанию: https://eclipse-sim.ru/swagger/doc.json -> openapi.yaml в корне репо.
"""
import json
import sys
import urllib.request

import yaml

HTTP_METHODS = {"get", "post", "put", "patch", "delete", "head", "options"}


def convert_ref(obj):
    if isinstance(obj, dict):
        if "$ref" in obj and isinstance(obj["$ref"], str):
            ref = obj["$ref"]
            if ref.startswith("#/definitions/"):
                return {"$ref": "#/components/schemas/" + ref.split("/")[-1]}
            if ref.startswith("#/parameters/"):
                return {"$ref": "#/components/parameters/" + ref.split("/")[-1]}
            return obj
        return {k: convert_ref(v) for k, v in obj.items()}
    if isinstance(obj, list):
        return [convert_ref(x) for x in obj]
    return obj


def convert_operation(op):
    out = {}
    for key, value in op.items():
        if key in ("parameters", "responses", "consumes", "produces"):
            continue
        out[key] = value

    params = []
    body_param = None
    for p in op.get("parameters", []):
        if p.get("in") == "body":
            body_param = p
            continue
        if p.get("in") == "formData":
            continue
        params.append(convert_ref(p))
    if params:
        out["parameters"] = params

    responses = {}
    produces = op.get("produces") or ["application/json"]
    media = produces[0] if produces else "application/json"
    for code, resp in op.get("responses", {}).items():
        entry = {"description": resp.get("description", "")}
        if "schema" in resp:
            entry["content"] = {media: {"schema": convert_ref(resp["schema"])}}
        responses[code] = entry
    out["responses"] = responses or {"200": {"description": "OK"}}

    if body_param is not None:
        out["requestBody"] = {
            "required": bool(body_param.get("required")),
            "content": {media: {"schema": convert_ref(body_param.get("schema", {}))}},
        }
    return out


def convert(doc):
    host = doc.get("host") or "eclipse-sim.ru"
    scheme = (doc.get("schemes") or ["https"])[0]
    base = "{}://{}{}".format(scheme, host, doc.get("basePath", "/"))
    out = {
        "openapi": "3.1.0",
        "info": doc.get("info", {"title": "API", "version": "1.0"}),
        "servers": [{"url": base}],
        "paths": {},
    }
    for path, methods in doc.get("paths", {}).items():
        entry = {}
        for method, op in methods.items():
            if method.lower() in HTTP_METHODS and isinstance(op, dict):
                entry[method.lower()] = convert_operation(op)
        if entry:
            out["paths"][path] = entry

    components = {}
    if doc.get("definitions"):
        components["schemas"] = convert_ref(doc["definitions"])
    sec = doc.get("securityDefinitions")
    if sec:
        converted = {}
        for name, s in sec.items():
            if s.get("type") == "apiKey":
                converted[name] = {
                    "type": "apiKey",
                    "name": s.get("name"),
                    "in": s.get("in"),
                }
        if converted:
            components["securitySchemes"] = converted
    if components:
        out["components"] = components
    return out


def load(source):
    if source.startswith("http://") or source.startswith("https://"):
        with urllib.request.urlopen(source, timeout=30) as r:
            return json.loads(r.read().decode())
    with open(source, encoding="utf-8") as f:
        return json.load(f)


if __name__ == "__main__":
    src = sys.argv[1] if len(sys.argv) > 1 else "https://eclipse-sim.ru/swagger/doc.json"
    dst = sys.argv[2] if len(sys.argv) > 2 else "openapi.yaml"
    data = convert(load(src))
    with open(dst, "w", encoding="utf-8") as f:
        yaml.safe_dump(data, f, allow_unicode=True, sort_keys=False)
    print(f"written {dst}: {len(data['paths'])} paths, {len(data.get('components', {}).get('schemas', {}))} schemas")
