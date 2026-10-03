"""Production-bundle regression with mocked APIs; never writes to a real backend.

Run npm run build first, then python tests/browser.py --channel chrome.
Requires Python Playwright and an installed Chrome/Edge (or Playwright Chromium).
"""

import argparse
import json
import threading
from functools import partial
from http.server import SimpleHTTPRequestHandler, ThreadingHTTPServer
from pathlib import Path
from urllib.parse import parse_qs, unquote, urlparse

from playwright.sync_api import expect, sync_playwright


class SPAHandler(SimpleHTTPRequestHandler):
    def do_GET(self):
        if not Path(self.translate_path(self.path)).exists():
            self.path = "/index.html"
        super().do_GET()

    def log_message(self, *args):
        pass


def run(base_url, channel):
    category = {"id": 1, "categoryName": "技术"}
    tag = {"id": 1, "tagName": "回归"}
    article = {
        "id": 1, "title": "依赖升级回归文章", "createTime": "2026-10-03T09:00:00+08:00",
        "category": category, "tags": [tag], "views": 10, "words": 200,
        "readTime": 1, "commentEnabled": True, "appreciation": False,
        "content": '<h2 id="smoke-heading">正文标题</h2><p>正文内容 😀</p>'
        '<img src="/img/avatar.jpg" alt="测试配图">'
        '<pre><code class="language-mermaid">graph TD; A[开始]--&gt;B[完成]</code></pre>',
    }
    private = {**article, "id": 2, "title": "受保护回归文章", "privacy": True}
    submitted = []
    requests = []
    unexpected = []

    def intercept(route):
        request = route.request
        parsed = urlparse(request.url)
        path = unquote(parsed.path.removeprefix("/api/"))
        if not parsed.path.startswith("/api/"):
            if parsed.hostname == urlparse(base_url).hostname:
                route.continue_()
            elif parsed.hostname == "v1.hitokoto.cn":
                route.fulfill(json={"hitokoto": "回归测试", "from": "测试"})
            else:
                route.abort()
            return
        requests.append({"path": path, "method": request.method,
                         "headers": request.headers, "query": parse_qs(parsed.query)})
        data = {}
        code = 0
        status = 200
        if path == "site":
            data = {"siteInfo": {"blogName": "回归博客", "webTitleSuffix": " | 回归博客"},
                    "introduction": {"name": "测试作者", "avatar": "/img/avatar.jpg",
                                     "rollText": [], "favorites": []},
                    "categoryList": [category], "tagList": [tag]}
        elif path in ("blogs", "category/blogs", "tag/blogs"):
            data = {"list": [article, private], "total": 2, "pageSize": 10}
        elif path == "blog":
            article_id = parse_qs(parsed.query)["id"][0]
            data = {**(private if article_id == "2" else article), "privacy": False}
        elif path == "comments":
            data = {"allComment": 1, "closeComment": 0, "comments": {"total": 1, "list": [{
                "id": 1, "nickname": "测试读者", "createTime": article["createTime"],
                "content": '<img src=x onerror="window.__commentXss=1"> 😀 <b>评论文本</b>',
                "avatar": "/img/avatar.jpg", "replyComments": [],
            }]}}
        elif path == "comment":
            submitted.append(request.post_data_json)
        elif path == "checkBlogPassword":
            assert request.post_data_json == {"blogId": 2, "password": "test-password"}
            data = "smoke-password-token"
        elif path == "searchBlog":
            data = [article]
        elif path == "archives":
            data = {"count": 2, "blogMap": {"2026-10": [article, private]}}
        elif path == "moments":
            data = {"list": [{"id": 1, "content": "回归动态", "isPublished": True,
                              "createTime": article["createTime"], "likes": 0}], "total": 1}
        elif path == "moment/like/1":
            pass
        elif path == "friends":
            data = {"friendList": [], "friendInfo": {"content": "友链说明", "commentEnabled": False}}
        elif path == "about":
            data = {"title": "关于回归博客", "content": "关于内容", "commentEnabled": "false"}
        elif path == "docs/tree":
            data = [{"title": "测试文档", "path": "test.md", "type": "file"}]
        elif path == "docs/content":
            data = {"path": "test.md", "title": "测试文档", "content": article["content"]}
        else:
            unexpected.append(path)
            code, status = 7, 404
        route.fulfill(status=status, json={"code": code, "data": data, "msg": "操作成功"},
                      headers={"identification": "smoke-visitor"})

    with sync_playwright() as playwright:
        browser = playwright.chromium.launch(headless=True, **({"channel": channel} if channel else {}))
        context = browser.new_context(viewport={"width": 1440, "height": 1000})
        context.route("**/*", intercept)
        page = context.new_page()
        errors = []
        page.on("pageerror", lambda error: errors.append(str(error)))
        try:
            page.goto(base_url + "/home")
            page.wait_for_load_state("networkidle")
            expect(page.locator(".article-title").first).to_have_text(article["title"])
            expect(page.locator(".article-meta").first).to_contain_text("2026-10-03")
            assert page.evaluate("localStorage.getItem('identification')") == "smoke-visitor"
            page.locator(".article-title a").first.click()
            expect(page).to_have_url(base_url + "/blog/1")
            expect(page.locator(".blog-title")).to_have_text(article["title"])
            expect(page.locator(".markdown-diagram img")).to_be_visible(timeout=15000)
            expect(page.locator(".comment > .content > .text")).to_contain_text("<img")
            assert page.locator(".comment > .content > .text img").count() == 0
            assert page.evaluate("window.__commentXss") is None
            page.locator('.js-toc-content img[alt="测试配图"]').click()
            expect(page.locator(".viewer-container.viewer-in")).to_be_visible()
            page.wait_for_function("document.querySelector('.js-toc-content').$viewer.viewed")
            page.locator(".viewer-button").click()
            expect(page.locator(".viewer-container")).to_be_hidden()

            page.get_by_role("button", name="发表评论", exact=True).click()
            expect(page.get_by_text("请输入评论昵称", exact=True)).to_be_visible()
            assert not submitted
            page.get_by_placeholder("昵称（必填）").fill("回归读者")
            page.get_by_placeholder("邮箱（必填）").fill("reader@example.com")
            page.get_by_placeholder("评论千万条，友善第一条").fill("升级后评论 😀")
            # The real submit button deliberately throttles clicks for three seconds.
            page.wait_for_timeout(3100)
            with page.expect_response(lambda response: urlparse(response.url).path == "/api/comment"):
                page.get_by_role("button", name="发表评论", exact=True).click(timeout=10000)
            assert submitted[-1]["content"] == "升级后评论 😀"
            assert submitted[-1]["nickname"] == "回归读者"
            assert submitted[-1]["blogId"] == 1
            assert submitted[-1]["page"] == 0

            page.locator('a.item[href="/home"]').click()
            page.locator(".article-title a").nth(1).click()
            dialog = page.get_by_role("dialog")
            expect(dialog).to_be_visible()
            dialog.locator("input").fill("test-password")
            dialog.get_by_role("button", name="确 定").click()
            expect(page).to_have_url(base_url + "/blog/2")
            expect(page.locator(".blog-title")).to_have_text(private["title"])
            assert page.evaluate("localStorage.getItem('blog2')") == "smoke-password-token"
            assert any(item["path"] == "blog" and item["query"].get("id") == ["2"]
                       and item["headers"].get("authorization") == "smoke-password-token" for item in requests)

            page.locator('a.item[href="/home"]').click()
            page.get_by_placeholder("Search...").fill("回归")
            expect(page.locator(".m-search-item .title")).to_have_text(article["title"])
            page.locator(".m-search-item .title").click()
            expect(page).to_have_url(base_url + "/blog/1")
            page.locator(".header-category").click()
            expect(page).to_have_url(base_url + "/category/%E6%8A%80%E6%9C%AF")
            expect(page.locator(".article-title").first).to_have_text(article["title"])
            page.locator(".article-tags a").first.click()
            expect(page).to_have_url(base_url + "/tag/%E5%9B%9E%E5%BD%92")
            expect(page.get_by_role("heading", name="标签 回归 下的文章", exact=True)).to_be_visible()

            for path, marker in [("archives", "文章归档"), ("moments", "我的动态"),
                                 ("friends", "小伙伴们"), ("about", "关于回归博客"), ("docs", "文档目录")]:
                page.locator(f'a.item[href="/{path}"]').click()
                expect(page.get_by_text(marker, exact=True)).to_be_visible()
                page.wait_for_load_state("networkidle")
                if path == "moments":
                    with page.expect_response(lambda response: "/moment/like/1" in response.url):
                        page.locator(".moment .extra a").click()
                    assert page.evaluate("localStorage.getItem('likeMomentIds')") == "[1]"
                if path == "docs":
                    expect(page.locator(".markdown-diagram img")).to_be_visible(timeout=15000)
            assert any(item["headers"].get("identification") == "smoke-visitor" for item in requests)
            assert not unexpected, f"Unexpected API calls: {unexpected}"
            assert not errors, f"Browser errors: {errors}"
            print("PASS: routes, dates, requests/headers, comments/escaping, password, search, image viewer, Mermaid, docs, likes")
        finally:
            browser.close()


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--channel", default="chrome", help="chrome, msedge, or empty for Playwright Chromium")
    parser.add_argument("--dist", type=Path, default=Path(__file__).resolve().parents[1] / "dist")
    args = parser.parse_args()
    if not (args.dist / "index.html").is_file():
        parser.error("Build output missing; run npm run build first")
    server = ThreadingHTTPServer(("127.0.0.1", 0), partial(SPAHandler, directory=str(args.dist.resolve())))
    threading.Thread(target=server.serve_forever, daemon=True).start()
    try:
        run(f"http://127.0.0.1:{server.server_port}", args.channel)
    finally:
        server.shutdown()
        server.server_close()
