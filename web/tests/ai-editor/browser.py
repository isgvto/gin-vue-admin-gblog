"""真实组件浏览器回归入口：文字聊天、配图、章节、模型管理和 Markdown 安全。"""
from pathlib import Path
import runpy
runpy.run_path(str(Path(__file__).with_name("chat_browser.py")), run_name="__main__")
