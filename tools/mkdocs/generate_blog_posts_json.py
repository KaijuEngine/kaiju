from __future__ import annotations

import json
from pathlib import Path
from typing import Any


POSTS_DIR = Path("blog/posts")
OUTPUT_FILE = Path("blog/posts.json")


def _docs_dir(config: dict[str, Any] | None = None) -> Path:
    if config is None:
        return Path("docs")
    return Path(config.get("docs_dir", "docs"))


def _read_existing_categories(output_path: Path) -> dict[str, str]:
    if not output_path.exists():
        return {}

    try:
        posts = json.loads(output_path.read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError):
        return {}

    categories: dict[str, str] = {}
    for post in posts:
        if not isinstance(post, dict):
            continue
        url = post.get("url")
        category = post.get("category")
        if isinstance(url, str) and isinstance(category, str):
            categories[url] = category
    return categories


def _front_matter(path: Path) -> dict[str, Any]:
    text = path.read_text(encoding="utf-8")
    if not text.startswith("---"):
        return {}

    parts = text.split("---", 2)
    if len(parts) < 3:
        return {}

    meta: dict[str, Any] = {}
    for line in parts[1].splitlines():
        stripped = line.strip()
        if not stripped or stripped.startswith("#") or ":" not in stripped:
            continue

        key, value = stripped.split(":", 1)
        value = value.strip().strip("'\"")
        if key.strip() == "tags":
            meta[key.strip()] = [tag.strip() for tag in value.split(",") if tag.strip()]
        else:
            meta[key.strip()] = value
    return meta


def _category(meta: dict[str, Any], url: str, existing_categories: dict[str, str]) -> str:
    if isinstance(meta.get("category"), str):
        return meta["category"]

    if url in existing_categories:
        return existing_categories[url]

    tags = meta.get("tags")
    if isinstance(tags, str):
        first_tag = tags.split(",", 1)[0].strip()
        if first_tag:
            return first_tag.title()
    if isinstance(tags, list) and tags:
        return str(tags[0]).title()

    return "Update"


def generate_posts_json(config: dict[str, Any] | None = None) -> list[dict[str, Any]]:
    docs_dir = _docs_dir(config)
    posts_dir = docs_dir / POSTS_DIR
    output_path = docs_dir / OUTPUT_FILE
    existing_categories = _read_existing_categories(output_path)
    posts: list[dict[str, Any]] = []

    for path in sorted(posts_dir.glob("*.md")):
        meta = _front_matter(path)
        url = f"posts/{path.stem}/"
        post: dict[str, Any] = {
            "title": str(meta.get("title") or path.stem.replace("_", " ").title()),
            "description": str(meta.get("description") or ""),
            "category": _category(meta, url, existing_categories),
            "date": str(meta.get("date") or ""),
            "url": url,
        }

        if meta.get("tags"):
            post["tags"] = meta["tags"]
        if meta.get("image"):
            post["image"] = str(meta["image"])

        posts.append(post)

    posts.sort(key=lambda post: post.get("date", ""), reverse=True)
    output_path.write_text(json.dumps(posts, indent=2) + "\n", encoding="utf-8")
    return posts


def on_config(config: dict[str, Any]) -> dict[str, Any]:
    generate_posts_json(config)
    return config


def on_pre_build(config: dict[str, Any]) -> None:
    generate_posts_json(config)


if __name__ == "__main__":
    generate_posts_json()
