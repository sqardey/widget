// ============================================================
//  Расписание — виджет для Scriptable (стиль «таймлайн»)
// ============================================================

// Ключ пользователя. Сервер подставляет сюда персональный ключ вместо {{KEY}}
const KEY = "{{API_KEY}}";

// ------------------------------------------------------------
//  Цвета
// ------------------------------------------------------------
const C = {
  bg1: new Color("#14151C"),
  bg2: new Color("#08080C"),
  text: Color.white(),
  muted: new Color("#8A8A96"),
  dim: new Color("#4A4A55"),
  red: new Color("#FF5A4F"),
  green: new Color("#34D399"),
  blue: new Color("#5AA9FF")
};

// ------------------------------------------------------------
//  Данные
// ------------------------------------------------------------
const FORCE = (args.queryParameters && args.queryParameters.force === "1") ? "&force=1" : "";
const API_URL = "https://widget.sqardey.xyz/api/widget?key=" + encodeURIComponent(KEY) + FORCE;

async function loadData() {
  try {
    let req = new Request(API_URL);
    req.timeoutInterval = 8;
    let json = await req.loadJSON();
    let status = req.response ? req.response.statusCode : 200;
    if (status === 401) return { data: null, error: "Неверный ключ" };
    if (status === 502) return { data: null, error: "Журнал недоступен" };
    if (status >= 400) return { data: null, error: "Ошибка сервера (" + status + ")" };
    return { data: json, error: null };
  } catch (e) {
    return { data: null, error: "Нет связи" };
  }
}

// ------------------------------------------------------------
//  Хелперы
// ------------------------------------------------------------
function toMin(t) {
  if (!t || t.indexOf(":") < 0) return null;
  let p = t.split(":");
  return parseInt(p[0], 10) * 60 + parseInt(p[1], 10);
}

// Сколько минут осталось до конца текущей пары
function minutesLeft(cur) {
  let e = toMin(cur.end_time);
  if (e === null) return null;
  let now = new Date();
  let n = now.getHours() * 60 + now.getMinutes();
  return Math.max(0, e - n);
}

function fmtLeft(m) {
  if (m === null) return "";
  if (m >= 60) return "ещё " + Math.floor(m / 60) + " ч " + (m % 60) + " мин";
  return "ещё " + m + " мин";
}

// Сокращение названия: «Теория вероятностей» → «ТВ», «Базы данных» → «БД»
function abbr(name) {
  name = (name || "").trim();
  if (name.length <= 10) return name;
  let words = name.split(/\s+/).filter(w => w.length > 2);
  if (words.length >= 2) {
    return words.slice(0, 4).map(w => w[0].toUpperCase()).join("");
  }
  return name.slice(0, 9) + ".";
}

function txt(stack, str, size, weight, color, lines) {
  let t = stack.addText(str);
  t.font = weight === "bold" ? Font.boldSystemFont(size)
    : weight === "semi" ? Font.semiboldSystemFont(size)
    : Font.mediumSystemFont(size);
  t.textColor = color;
  if (lines) t.lineLimit = lines;
  return t;
}

// Одна строка таймлайна:  [время] ▍ [название / метка]            [справа]
function row(parent, o, compact) {
  let r = parent.addStack();
  r.centerAlignContent();
  r.spacing = 8;

  // Время
  let tc = r.addStack();
  tc.layoutVertically();
  let t1 = tc.addText(o.start);
  t1.font = Font.boldMonospacedSystemFont(compact ? 13 : 15);
  t1.textColor = C.text;
  t1.lineLimit = 1;
  if (o.end) {
    let t2 = tc.addText(o.end);
    t2.font = Font.mediumMonospacedSystemFont(compact ? 9 : 10);
    t2.textColor = C.muted;
    t2.lineLimit = 1;
  }

  // Цветная полоска
  let bar = r.addStack();
  bar.size = new Size(3, compact ? 30 : 34);
  bar.backgroundColor = o.accent;
  bar.cornerRadius = 1.5;

  // Текст
  let body = r.addStack();
  body.layoutVertically();
  body.spacing = 2;
  let title = txt(body, o.title, compact ? 12 : 14, "bold", C.text, 1);
  title.minimumScaleFactor = 0.75;
  txt(body, o.label, 9.5, "bold", o.accent, 1);

  // Справа
  if (!compact && o.right) {
    r.addSpacer();
    txt(r, o.right, 11, "semi", o.accent, 1);
  }
  return r;
}

// ------------------------------------------------------------
//  Обычный виджет (домашний экран)
// ------------------------------------------------------------
function buildMain(data, error, compact) {
  let widget = new ListWidget();
  let grad = new LinearGradient();
  grad.colors = [C.bg1, C.bg2];
  grad.locations = [0, 1];
  widget.backgroundGradient = grad;
  widget.setPadding(12, 14, 8, 14);
  widget.url = TAP_URL;

  if (!data) {
    widget.addSpacer();
    let e = widget.addText("⚠️ " + error);
    e.textColor = C.red;
    e.font = Font.boldSystemFont(14);
    let s = widget.addText("Проверь ключ, сервер или сеть");
    s.textColor = C.muted;
    s.font = Font.systemFont(11);
    widget.addSpacer();
    return widget;
  }

  let exam = (data.has_exam && data.exam) ? data.exam : null;
  let cur = data.current_class;
  let nxt = data.next_class;
  let rows = [];

  if (exam) {
    rows.push({
      accent: C.red,
      start: exam.start_time || "—",
      end: null,
      title: exam.name_subject || "Экзамен",
      label: "ЭКЗАМЕН" + (exam.audience ? " · " + exam.audience + " ауд." : ""),
      right: "сегодня"
    });
  }
  if (cur) {
    rows.push({
      accent: C.green,
      start: cur.start_time || "••:••",
      end: cur.end_time,
      title: cur.name_subject,
      label: "СЕЙЧАС · " + cur.audience + " ауд.",
      right: fmtLeft(minutesLeft(cur))
    });
  }
  if (nxt) {
    rows.push({
      accent: C.blue,
      start: nxt.start_time,
      end: nxt.end_time,
      title: nxt.name_subject,
      label: "ДАЛЬШЕ · " + nxt.audience + " ауд.",
      right: ""
    });
  }

  if (rows.length === 0) {
    widget.addSpacer();
    let a = widget.addText("Пар больше нет 🎉");
    a.font = Font.boldSystemFont(15);
    a.textColor = C.text;
    a.centerAlignText();
    let b = widget.addText("Можно отдыхать");
    b.font = Font.mediumSystemFont(11);
    b.textColor = C.muted;
    b.centerAlignText();
    widget.addSpacer();
  } else {
    widget.addSpacer(2);
    rows.forEach((o, i) => {
      row(widget, o, compact);
      if (i < rows.length - 1) widget.addSpacer(compact ? 6 : 7);
    });
    widget.addSpacer();
  }

  let footer = widget.addText("обновлено " + (data.updated_at || ""));
  footer.textColor = C.dim;
  footer.font = Font.systemFont(8);
  footer.rightAlignText();
  return widget;
}

// ------------------------------------------------------------
//  Экран блокировки
// ------------------------------------------------------------
function buildLock(data, error, kind) {
  let w = new ListWidget();
  w.url = TAP_URL;

  let cur = data && data.current_class;
  let nxt = data && data.next_class;
  let exam = (data && data.has_exam && data.exam) ? data.exam : null;

  if (kind === "inline") {
    // Одна строка: только одна пара (текущая, иначе следующая) + начало и конец
    let line;
    if (!data) {
      line = "⚠️ " + error;
    } else {
      let c = cur || nxt;
      line = c ? abbr(c.name_subject) + " " + c.start_time + "–" + c.end_time : "Пар нет 🎉";
    }
    w.addText(line);
  } else if (kind === "rectangular") {
    let lines = [];
    if (!data) {
      lines.push("⚠️ " + error);
    } else {
      if (exam) lines.push("🔥 " + abbr(exam.name_subject) + " " + (exam.start_time || ""));
      lines.push(cur ? "● " + abbr(cur.name_subject) + " до " + cur.end_time : "● Сейчас окно");
      lines.push(nxt ? "→ " + abbr(nxt.name_subject) + " " + nxt.start_time : "→ Дальше пар нет");
    }
    lines.forEach((l, i) => {
      let t = w.addText(l);
      t.font = i === 0 ? Font.boldSystemFont(13) : Font.systemFont(12);
      t.lineLimit = 1;
    });
  } else {
    // circular
    let top = w.addText(cur ? "до" : "в");
    top.font = Font.systemFont(9);
    top.centerAlignText();
    let val = !data ? "⚠️" : cur ? cur.end_time : (nxt ? nxt.start_time : "—");
    let t = w.addText(val);
    t.font = Font.boldSystemFont(14);
    t.centerAlignText();
  }
  return w;
}

// ------------------------------------------------------------
//  Запуск
// ------------------------------------------------------------
// Нажатие на виджет запускает скрипт с force=1 (принудительное обновление)
const TAP_URL = "scriptable:///run/" + encodeURIComponent(Script.name()) + "?force=1";

const result = await loadData();
const family = String(config.widgetFamily || "medium");
const lockKind = family.replace("accessory", "").toLowerCase();
const isLock = ["inline", "rectangular", "circular"].includes(lockKind);

const out = isLock
  ? buildLock(result.data, result.error, lockKind)
  : buildMain(result.data, result.error, family === "small");

if (config.runsInWidget) {
  Script.setWidget(out);
} else {
  out.presentMedium();
}
Script.complete();