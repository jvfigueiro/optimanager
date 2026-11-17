#!/usr/bin/env python3
"""Seed a running OptiManager instance with demo data from private CSV exports.

The CSV files contain personal data and must NEVER be committed. This script
contains no personal data itself; it only reads the files passed by argument and
writes through the public HTTP API (so schema rules and numbering are respected).

Usage:
  scripts/seed_demo.py --base-url http://localhost:8686 \
      --clientes path/clientes.csv --servicos path/servicos.csv
"""
import argparse
import csv
import json
import sys
import urllib.error
import urllib.request

ADMIN_NAME = "Administrador Demo"
ADMIN_EMAIL = "admin@optimanager.local"
ADMIN_PASSWORD = "demo1234"


def read_csv(path):
    raw = open(path, "rb").read()
    for enc in ("utf-8-sig", "cp1252", "latin-1"):
        try:
            text = raw.decode(enc)
            break
        except UnicodeDecodeError:
            continue
    return list(csv.reader(text.splitlines()))


def call(base, method, path, body=None, token=None):
    data = json.dumps(body).encode() if body is not None else None
    req = urllib.request.Request(base + path, data=data, method=method)
    req.add_header("Content-Type", "application/json")
    if token:
        req.add_header("Authorization", "Bearer " + token)
    try:
        with urllib.request.urlopen(req) as resp:
            payload = resp.read()
            return resp.status, (json.loads(payload) if payload else None)
    except urllib.error.HTTPError as e:
        return e.code, e.read().decode(errors="replace")


def iso_date(br_date, default_year="2021"):
    parts = br_date.strip().split("/")
    if len(parts) == 2:  # day/month without year: assume the export's year
        parts.append(default_year)
    d, m, y = parts
    return f"{y}-{m.zfill(2)}-{d.zfill(2)}"


def parse_money(txt):
    s = txt.replace("R$", "").strip()
    if not s:
        return 0.0
    if "," in s:  # Brazilian format: 1.500,00
        s = s.replace(".", "").replace(",", ".")
    elif s.count(".") > 1:  # 1.500.00 -> thousands dots + decimal dot
        head, _, tail = s.rpartition(".")
        s = head.replace(".", "") + "." + tail
    return float(s)


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--base-url", default="http://localhost:8686")
    ap.add_argument("--clientes", required=True)
    ap.add_argument("--servicos", required=True)
    args = ap.parse_args()
    base = args.base_url.rstrip("/")

    # Admin account (setup on first run, otherwise login)
    status, resp = call(base, "POST", "/api/auth/setup",
                        {"nome": ADMIN_NAME, "email": ADMIN_EMAIL, "senha": ADMIN_PASSWORD})
    if status not in (200, 201):
        status, resp = call(base, "POST", "/api/auth/login",
                            {"email": ADMIN_EMAIL, "senha": ADMIN_PASSWORD})
    if status != 200 and status != 201:
        sys.exit(f"auth failed: {status} {resp}")
    token = resp["token"]

    # Clients: columns [_, desde, nome, cpf/cnpj, _]
    n_cli = 0
    for r in read_csv(args.clientes)[3:]:
        if len(r) < 3 or not r[2].strip():
            continue
        body = {"nome": r[2].strip(), "doc": (r[3].strip() if len(r) > 3 else ""),
                "contato": "", "desde_ano": int(r[1]) if r[1].strip().isdigit() else 0}
        st, _ = call(base, "POST", "/api/clientes", body, token)
        n_cli += st == 201

    # Services: columns [_, data, status, cliente, equipamento, servico, valor, forma, _]
    rows = [r for r in read_csv(args.servicos)[3:] if len(r) > 7 and r[1].strip()]
    rows.sort(key=lambda r: iso_date(r[1]))  # oldest first -> ascending order numbers
    n_os = 0
    for r in rows:
        data = iso_date(r[1])
        servico = r[5].strip()
        is_vt = servico.upper().startswith("VT - ")
        body = {
            "data_emissao": data, "hora_emissao": "09:00",
            "data_entrega": data, "hora_entrega": "17:00",
            "status": "Concluído" if r[2].strip().lower() == "pronto" else r[2].strip(),
            "cliente_nome": r[3].strip(),
            "equipamento": r[4].strip() or "N/A",
            "servico": servico,
            "valor_total": parse_money(r[6]),
            "pagamento_status": "Pago",
            "pagamento_forma": r[7].strip() or "Dinheiro",
            "is_vt": is_vt,
        }
        st, out = call(base, "POST", "/api/ordens", body, token)
        if st == 201:
            n_os += 1
        else:
            print(f"warn: order skipped ({st}): {out}", file=sys.stderr)

    print(f"seeded: {n_cli} clients, {n_os} orders")
    print(f"login: {ADMIN_EMAIL} / {ADMIN_PASSWORD}")


if __name__ == "__main__":
    main()
