@echo off
cd /d "%~dp0"
where py >nul 2>&1
if errorlevel 1 (
  echo Python launcher not found. Install Python 3.11+ first.
  pause
  exit /b 1
)
if not exist .venv\Scripts\python.exe (
  py -3 -m venv .venv
  .venv\Scripts\python.exe -m pip install -r requirements.txt
)
if not exist .env (
  copy .env.example .env >nul
  echo Edit .env with your Telegram Bot Token and Chat ID first.
  pause
  exit /b 1
)
.venv\Scripts\python.exe monitor.py
pause
