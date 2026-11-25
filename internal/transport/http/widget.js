// Виджет расписания IT-TOP для Scriptable. Версия 2.
const KEY = "{{API_KEY}}";
const BASE_URL = "{{BASE_URL}}";

// Предпросмотр внутри приложения Scriptable: "small", "medium" или "large".
const PREVIEW = "medium";

// Тема: "auto" (как в системе), "light" или "dark" (принудительно).
const THEME = "auto";

// Тап по виджету запускает скрипт с force=1: сервер обновляет данные из журнала (не чаще раза в час)
const FORCE = (args.queryParameters && args.queryParameters.force === "1") ? "&force=1" : "";
const API_URL = BASE_URL + "/api/widget?key=" + encodeURIComponent(KEY) + FORCE;

// 401 -> auth, 502 (журнал не ответил) -> journal, остальное -> server, нет сети -> net
async function loadData() {
  try {
    let req = new Request(API_URL);
    req.timeoutInterval = 8;
    let text = await req.loadString();
    let status = req.response ? req.response.statusCode : 200;
    if (status === 401) return { data: null, error: "auth" };
    if (status === 502) return { data: null, error: "journal" };
    if (status >= 400) return { data: null, error: "server" };
    return { data: JSON.parse(text), error: null };
  } catch (e) {
    return { data: null, error: "net" };
  }
}

// Цвета задаются парой (светлая и тёмная тема). Какую показать, решает сама iOS
// в момент отрисовки, поэтому виджет не «залипает» в чужой теме.
const PD = {
  bgTop: "#1B1E36", bgBottom: "#0F1120",
  text: "#F2F3FA", sub: "#9A9EC0", line: "#FFFFFF", lineA: 0.12,
  now: "#FFB547", next: "#97A3FF", exam: "#FF7A6B", pill: "#FFFFFF", pillA: 0.12, track: "#FFFFFF", trackA: 0.14
};
const PL = {
  bgTop: "#F5F6FC", bgBottom: "#E8EAF6",
  text: "#171A30", sub: "#5E6288", line: "#171A30", lineA: 0.12,
  now: "#B86E00", next: "#4453D8", exam: "#CC3B2B", pill: "#171A30", pillA: 0.08, track: "#171A30", trackA: 0.12
};
const P = {};
Object.keys(PD).forEach(k => { P[k] = k; });
// col("text") или col("text", 0.5) или col("pill", "pillA")
const col = (k, a) => {
  const pick = T => new Color(T[k], a === undefined ? 1 : (typeof a === "string" ? T[a] : a));
  if (THEME === "light") return pick(PL);
  if (THEME === "dark") return pick(PD);
  return Color.dynamic(pick(PL), pick(PD));
};
const themed = (l, d) => THEME === "light" ? l : (THEME === "dark" ? d : Color.dynamic(l, d));

const SW = Device.screenSize().width;
const FAMILY = config.runsInWidget ? config.widgetFamily : PREVIEW;
const IS_ACCESSORY = String(FAMILY).indexOf("accessory") === 0;
const IS_SMALL = FAMILY === "small";
const IS_LARGE = FAMILY === "large" || FAMILY === "extraLarge";

function parseHM(s) {
  const m = /(\d{1,2}):(\d{2})/.exec(s || "");
  if (!m) return null;
  return { h: +m[1], m: +m[2], text: ("0" + m[1]).slice(-2) + ":" + m[2] };
}
// Время пар приходит строкой «ЧЧ:ММ» в часовом поясе сервера и трактуется как локальное время телефона
function todayAt(t) {
  const d = new Date();
  d.setHours(t.h, t.m, 0, 0);
  return d;
}
function lessonTimes(cls) {
  const s = parseHM(cls.start_time), e = parseHM(cls.end_time);
  return {
    start: s, end: e,
    startDate: s ? todayAt(s) : null,
    endDate: e ? todayAt(e) : null,
    range: (s ? s.text : (cls.start_time || "")) + "–" + (e ? e.text : (cls.end_time || ""))
  };
}
function progressOf(t) {
  if (!t.startDate || !t.endDate) return 0;
  const total = t.endDate - t.startDate;
  if (total <= 0) return 0;
  return Math.max(0, Math.min(1, (new Date() - t.startDate) / total));
}

function mix(a, b, t) {
  const p = h => [1, 3, 5].map(i => parseInt(h.substr(i, 2), 16));
  const x = p(a), y = p(b);
  return "#" + x.map((v, i) => ("0" + Math.round(v + (y[i] - v) * t).toString(16)).slice(-2)).join("");
}

// Фон слегка подкрашивается цветом состояния (сейчас / дальше)
function setBackground(widget, tint) {
  const g = new LinearGradient();
  const top = (T, k) => new Color(tint ? mix(T.bgTop, T[tint], k) : T.bgTop);
  g.colors = [themed(top(PL, 0.10), top(PD, 0.18)), col(P.bgBottom)];
  g.locations = [0, 1];
  widget.backgroundGradient = g;
}

function addBar(parent, width, progress, color) {
  const H = 6;
  const track = parent.addStack();
  track.layoutHorizontally();
  track.size = new Size(width, H);
  track.cornerRadius = H / 2;
  track.backgroundColor = col(P.track, P.trackA);
  if (progress > 0) {
    const fill = track.addStack();
    fill.size = new Size(Math.max(H, width * progress), H);
    fill.cornerRadius = H / 2;
    fill.backgroundColor = color;
  }
  track.addSpacer();
}

// Шкала на всю доступную ширину: прогресс задан градиентом с резкой границей
function addFlexBar(parent, progress, color) {
  const H = 6;
  const p = Math.max(0, Math.min(1, progress));
  const track = parent.addStack();
  track.layoutHorizontally();
  track.size = new Size(0, H);
  track.cornerRadius = H / 2;
  const tr = col(P.track, P.trackA);
  const g = new LinearGradient();
  g.startPoint = new Point(0, 0.5);
  g.endPoint = new Point(1, 0.5);
  g.colors = [color, color, tr, tr];
  g.locations = [0, p, p, 1];
  track.backgroundGradient = g;
  track.addSpacer();
}

// «Аудитория 14» / «ауд. 14» -> «14»
function room(a) {
  return String(a || "").replace(/^\s*(аудитория|ауд\.?)\s*/i, "").trim();
}

// Иконка двери (из SVG), рисуется вручную. Цвет задаётся через tintColor.
function doorIcon() {
  const k = 4; // 24 -> 96 px
  const ctx = new DrawContext();
  ctx.size = new Size(96, 96);
  ctx.opaque = false;
  ctx.respectScreenScale = false;
  const pt = (x, y) => new Point(x * k, y * k);

  const f = new Path();
  f.move(pt(14.37, 4));
  f.addCurve(pt(17.41, 4.3), pt(15.98, 4), pt(16.79, 4));
  f.addCurve(pt(18.8, 5.68), pt(18.0, 4.55), pt(18.5, 5.0));
  f.addCurve(pt(19.1, 8.73), pt(19.1, 6.3), pt(19.1, 7.11));
  f.addLine(pt(19.1, 21.6));
  f.addLine(pt(14.37, 21.6));
  f.closeSubpath();
  ctx.setFillColor(new Color("#FFFFFF", 0.2));
  ctx.addPath(f);
  ctx.fillPath();

  ctx.setStrokeColor(Color.white());
  ctx.setLineWidth(1.5 * k);

  const h = new Path();
  h.move(pt(9.87, 12.48));
  h.addLine(pt(11.6, 12.12));
  ctx.addPath(h);
  ctx.strokePath();

  const a = new Path();
  a.move(pt(14.37, 21.32));
  a.addLine(pt(14.37, 7.26));
  a.addCurve(pt(13.76, 3.04), pt(14.37, 4.92), pt(14.37, 3.75));
  a.addCurve(pt(12.89, 2.4), pt(13.58, 2.75), pt(13.2, 2.55));
  a.addCurve(pt(8.69, 3.02), pt(12.04, 2.02), pt(10.92, 2.35));
  a.addCurve(pt(5.79, 4.26), pt(7.11, 3.49), pt(6.32, 3.72));
  a.addCurve(pt(5.26, 4.97), pt(5.6, 4.5), pt(5.4, 4.75));
  a.addCurve(pt(4.9, 8.1), pt(4.9, 5.64), pt(4.9, 6.46));
  a.addLine(pt(4.9, 21.32));
  ctx.addPath(a);
  ctx.strokePath();

  const g = new Path();
  g.move(pt(21, 21.61));
  g.addLine(pt(3, 21.61));
  ctx.addPath(g);
  ctx.strokePath();

  const r = new Path();
  r.move(pt(19.1, 21.32));
  r.addLine(pt(19.1, 8.73));
  r.addCurve(pt(18.8, 5.68), pt(19.1, 7.11), pt(19.1, 6.3));
  r.addCurve(pt(17.4, 4.3), pt(18.5, 5.0), pt(18.0, 4.55));
  r.addCurve(pt(14.37, 4), pt(16.8, 4.0), pt(15.98, 4.0));
  ctx.addPath(r);
  ctx.strokePath();

  return ctx.getImage();
}

function addExamBanner(parent, exam, size, w, h) {
  const box = parent.addStack();
  box.layoutHorizontally();
  box.centerAlignContent();
  box.size = new Size(w, h);
  box.backgroundColor = col(P.exam, 0.16);
  box.cornerRadius = h * 0.38;
  box.setPadding(0, 10, 0, 10);

  const sym = SFSymbol.named("graduationcap.fill");
  if (sym) {
    sym.applyFont(Font.boldSystemFont(size));
    const im = box.addImage(sym.image);
    im.imageSize = new Size(size * 1.35, size);
    im.applyFittingContentMode();
    im.tintColor = col(P.exam);
    box.addSpacer(6);
  }
  const lab = box.addText("ЭКЗАМЕН");
  lab.font = Font.heavyRoundedSystemFont(size * 0.9);
  lab.textColor = col(P.exam);
  lab.lineLimit = 1;
  box.addSpacer(8);
  const t = box.addText((exam && exam.name_subject) || "");
  t.font = Font.semiboldSystemFont(size);
  t.textColor = col(P.text);
  t.lineLimit = 1;
  t.minimumScaleFactor = 0.7;
  box.addSpacer();
  const rm = room(exam && exam.audience);
  const when = (exam && (exam.start_time || exam.end_time)) ? lessonTimes(exam).range : "";
  const tail = [when, rm ? "ауд. " + rm : ""].filter(Boolean).join(" · ") || "сегодня";
  const td = box.addText(tail);
  td.font = Font.mediumRoundedSystemFont(size * 0.9);
  td.textColor = col(P.exam, 0.85);
  td.lineLimit = 1;
  td.minimumScaleFactor = 0.7;
}

// Ширина содержимого малого виджета зависит от модели iPhone
function smallInnerWidth() {
  const w = Device.screenSize().width;
  const box = w <= 375 ? 141 : (w <= 393 ? 158 : 170);
  return box - 28 - 2;
}
const isTight = () => Device.screenSize().width <= 375;

function addDoorPill(parent, rm, s) {
  if (!rm) return;
  s = s || 1;
  const pill = parent.addStack();
  pill.layoutHorizontally();
  pill.centerAlignContent();
  pill.backgroundColor = col(P.pill, P.pillA);
  pill.cornerRadius = 8 * s;
  pill.setPadding(3 * s, 7 * s, 3 * s, 8 * s);
  const im = pill.addImage(doorIcon());
  im.imageSize = new Size(13 * s, 13 * s);
  im.tintColor = col(P.sub);
  pill.addSpacer(4 * s);
  const at = pill.addText(rm);
  at.font = Font.boldRoundedSystemFont(12 * s);
  at.textColor = col(P.text);
  at.lineLimit = 1;
  at.minimumScaleFactor = 0.5;
}

function addStatusHead(widget, label, accent, rm) {
  const head = widget.addStack();
  head.layoutHorizontally();
  head.centerAlignContent();
  const dot = head.addStack();
  dot.size = new Size(6, 6);
  dot.cornerRadius = 3;
  dot.backgroundColor = accent;
  head.addSpacer(5);
  const lab = head.addText(label);
  lab.font = Font.heavyRoundedSystemFont(10.5);
  lab.textColor = accent;
  lab.lineLimit = 1;
  head.addSpacer();
  addDoorPill(head, rm);
}

// Плашка пары, подкрашенная в цвет состояния.
// Компактная (detailed=false): «ДАЛЬШЕ  10:45 · 305» + предмет.
// Подробная (detailed=true): «СЕЙЧАС  до 20:18» + предмет + строка с аудиторией.
function addClassBox(parent, cls, now, tight, detailed) {
  const key = now ? P.now : P.next;
  const accent = col(key);
  const tm = lessonTimes(cls);
  const rm = room(cls.audience);

  const box = parent.addStack();
  box.layoutVertically();
  box.backgroundColor = col(key, 0.15);
  box.cornerRadius = 10;
  box.setPadding(5, 9, 5, 9);

  const r1 = box.addStack();
  r1.layoutHorizontally();
  r1.centerAlignContent();
  const l = r1.addText(now ? "СЕЙЧАС" : "ДАЛЬШЕ");
  l.font = Font.heavyRoundedSystemFont(9);
  l.textColor = accent;
  l.lineLimit = 1;
  l.minimumScaleFactor = 0.6;
  r1.addSpacer();
  let when = now
    ? "до " + (tm.end ? tm.end.text : (cls.end_time || ""))
    : (tm.start ? tm.start.text : (cls.start_time || ""));
  if (!detailed && rm) when += " · " + rm;
  const w = r1.addText(when);
  w.font = Font.boldRoundedSystemFont(11);
  w.textColor = col(P.text);
  w.lineLimit = 1;
  w.minimumScaleFactor = 0.6;

  box.addSpacer(2);
  const sb = box.addText(cls.name_subject || "");
  sb.font = Font.semiboldSystemFont(10.5);
  sb.textColor = col(P.text);
  sb.lineLimit = 1;
  sb.minimumScaleFactor = 0.7;

  if (detailed && rm) {
    box.addSpacer(2);
    const r3 = box.addStack();
    r3.layoutHorizontally();
    r3.centerAlignContent();
    const im = r3.addImage(doorIcon());
    im.imageSize = new Size(11, 11);
    im.tintColor = col(P.sub);
    r3.addSpacer(4);
    const rt = r3.addText("ауд. " + rm);
    rt.font = Font.mediumRoundedSystemFont(10);
    rt.textColor = col(P.sub);
    rt.lineLimit = 1;
    r3.addSpacer();
  }
}

function addNoMoreBox(parent) {
  const box = parent.addStack();
  box.layoutHorizontally();
  box.centerAlignContent();
  box.backgroundColor = col(P.pill, P.pillA);
  box.cornerRadius = 9;
  box.setPadding(6, 8, 6, 8);
  const l = box.addText("ДАЛЬШЕ");
  l.font = Font.heavyRoundedSystemFont(8.5);
  l.textColor = col(P.sub);
  l.lineLimit = 1;
  box.addSpacer();
  const t = box.addText("пар нет");
  t.font = Font.mediumRoundedSystemFont(10.5);
  t.textColor = col(P.sub);
  t.lineLimit = 1;
}

function addEnds(parent, a, b, s) {
  s = s || 1;
  const ends = parent.addStack();
  ends.layoutHorizontally();
  const t1 = ends.addText(a);
  t1.font = Font.mediumRoundedSystemFont(10 * s);
  t1.textColor = col(P.sub);
  ends.addSpacer();
  const t2 = ends.addText(b);
  t2.font = Font.mediumRoundedSystemFont(10 * s);
  t2.textColor = col(P.sub);
}

// День с экзаменом: сверху плашка экзамена, снизу текущая или следующая пара
function buildSmallExam(widget, exam, cur, nxt, tight) {
  widget.setPadding(12, 12, 11, 12);
  const accent = col(P.exam);
  const hasClass = Boolean(cur || nxt);

  const H = SW <= 375 ? 141 : (SW <= 393 ? 158 : 170);
  const boxH = H - 23 - (hasClass ? 52 : 30) - 12;

  const box = widget.addStack();
  box.layoutVertically();
  box.size = new Size(0, boxH);
  box.backgroundColor = col(P.exam, 0.16);
  box.cornerRadius = 10;
  box.setPadding(7, 9, 8, 9);

  const head = box.addStack();
  head.layoutHorizontally();
  head.centerAlignContent();
  const sym = SFSymbol.named("graduationcap.fill");
  if (sym) {
    sym.applyFont(Font.boldSystemFont(12));
    const im = head.addImage(sym.image);
    im.imageSize = new Size(16, 12);
    im.applyFittingContentMode();
    im.tintColor = accent;
    head.addSpacer(4);
  }
  const lab = head.addText("ЭКЗАМЕН");
  lab.font = Font.heavyRoundedSystemFont(10);
  lab.textColor = accent;
  lab.lineLimit = 1;
  lab.minimumScaleFactor = 0.7;
  head.addSpacer();
  const rm = room(exam.audience);
  if (rm) {
    addDoorPill(head, rm);
  } else {
    const td = head.addText("сегодня");
    td.font = Font.mediumRoundedSystemFont(10);
    td.textColor = col(P.exam, 0.85);
    td.lineLimit = 1;
    td.minimumScaleFactor = 0.7;
  }

  box.addSpacer();
  const name = box.addText(exam.name_subject || "Экзамен");
  name.font = Font.boldSystemFont(hasClass ? 16 : 18);
  name.textColor = col(P.text);
  name.lineLimit = hasClass ? 3 : 5;
  name.minimumScaleFactor = 0.7;

  if (exam.start_time || exam.end_time) {
    box.addSpacer(2);
    const t = box.addText(lessonTimes(exam).range);
    t.font = Font.mediumRoundedSystemFont(10.5);
    t.textColor = col(P.exam, 0.85);
    t.lineLimit = 1;
  }
  box.addSpacer();

  widget.addSpacer(8);

  if (cur) addClassBox(widget, cur, true, tight, true);
  else if (nxt) addClassBox(widget, nxt, false, tight, true);
  else addNoMoreBox(widget);
}

function buildSmallBoth(widget, cur, nxt, tight) {
  widget.setPadding(10, 14, 10, 14);
  const tm = lessonTimes(cur);
  const accent = col(P.now);

  addStatusHead(widget, "СЕЙЧАС", accent, room(cur.audience));
  widget.addSpacer(2);

  if (tm.endDate) {
    const d = widget.addDate(tm.endDate);
    d.applyTimerStyle();
    d.font = Font.boldRoundedSystemFont(26);
    d.textColor = accent;
    d.lineLimit = 1;
    d.minimumScaleFactor = 0.6;
  }

  const subj = widget.addText(cur.name_subject || "");
  subj.font = Font.semiboldSystemFont(11.5);
  subj.textColor = col(P.text);
  subj.lineLimit = nxt ? 2 : 3;
  subj.minimumScaleFactor = 0.8;

  widget.addSpacer();

  addEnds(widget, tm.start ? tm.start.text : (cur.start_time || ""), tm.end ? tm.end.text : (cur.end_time || ""));
  widget.addSpacer(3);
  addFlexBar(widget, progressOf(tm), accent);
  widget.addSpacer(6);

  if (nxt) addClassBox(widget, nxt, false, tight, false);
  else addNoMoreBox(widget);
}

function buildSmall(widget, data) {
  widget.setPadding(13, 14, 12, 14);
  const cur = data.current_class, nxt = data.next_class;
  const exam = data.has_exam && data.exam ? data.exam : null;
  const tight = isTight();
  const W = smallInnerWidth();

  if (exam) {
    buildSmallExam(widget, exam, cur, nxt, tight);
    return;
  }
  if (cur) {
    buildSmallBoth(widget, cur, nxt, tight);
    return;
  }

  if (!nxt) {
    widget.addSpacer();
    const t = widget.addText("На сегодня всё");
    t.font = Font.boldRoundedSystemFont(22);
    t.textColor = col(P.text);
    t.minimumScaleFactor = 0.8;
    widget.addSpacer(2);
    const s = widget.addText("Пар больше нет");
    s.font = Font.systemFont(12);
    s.textColor = col(P.sub);
    widget.addSpacer();
    addFooter(widget, data, 9);
    return;
  }

  // Пара ещё не началась: крупное время начала и «через …»
  const accent = col(P.next);
  const tm = lessonTimes(nxt);

  addStatusHead(widget, "ДАЛЬШЕ", accent, room(nxt.audience));
  widget.addSpacer(5);

  const big = widget.addText(tm.start ? tm.start.text : (nxt.start_time || ""));
  big.font = Font.boldRoundedSystemFont(36);
  big.textColor = col(P.text);
  big.lineLimit = 1;
  if (tm.startDate && tm.startDate > new Date()) {
    const rel = widget.addDate(tm.startDate);
    rel.applyRelativeStyle();
    rel.font = Font.mediumSystemFont(10.5);
    rel.textColor = col(P.sub);
    rel.lineLimit = 1;
  }

  widget.addSpacer(4);
  const subj = widget.addText(nxt.name_subject || "");
  subj.font = Font.semiboldSystemFont(12.5);
  subj.textColor = col(P.text);
  subj.lineLimit = 3;
  subj.minimumScaleFactor = 0.8;

  widget.addSpacer();
  addEnds(widget, tm.start ? tm.start.text : (nxt.start_time || ""), tm.end ? tm.end.text : (nxt.end_time || ""));
  widget.addSpacer(3);
  addBar(widget, W, 0, accent);
}

// Размеры виджетов по ширине экрана iPhone (в пунктах)
const WIDE_W = SW <= 375 ? 321 : (SW <= 414 ? 338 : 364);
const MED_H = SW <= 375 ? 148 : (SW <= 414 ? 158 : 170);
const LRG_H = SW <= 375 ? 324 : (SW <= 414 ? 354 : 382);

// Карточка пары: статус + аудитория, крупный таймер/время, предмет, шкала.
// o: { w, h, s (масштаб), pv, ph, radius, big, compact, sub }
function addCard(parent, cls, now, o) {
  const key = now ? P.now : P.next;
  const accent = col(key);
  const tm = lessonTimes(cls);
  const rm = room(cls.audience);
  const s = o.s;

  const card = parent.addStack();
  card.layoutVertically();
  card.size = new Size(o.w, o.h);
  card.backgroundColor = col(key, 0.15);
  card.cornerRadius = o.radius;
  card.setPadding(o.pv, o.ph, o.pv, o.ph);

  const head = card.addStack();
  head.layoutHorizontally();
  head.centerAlignContent();
  const dot = head.addStack();
  dot.size = new Size(6, 6);
  dot.cornerRadius = 3;
  dot.backgroundColor = accent;
  head.addSpacer(5);
  const lab = head.addText(now ? "СЕЙЧАС" : "ДАЛЬШЕ");
  lab.font = Font.heavyRoundedSystemFont(10.5 * s);
  lab.textColor = accent;
  lab.lineLimit = 1;
  head.addSpacer();
  addDoorPill(head, rm, s);

  card.addSpacer(o.compact ? 1 : 4 * s);

  if (now && tm.endDate) {
    const d = card.addDate(tm.endDate);
    d.applyTimerStyle();
    d.font = Font.boldRoundedSystemFont(o.big);
    d.textColor = accent;
    d.lineLimit = 1;
    d.minimumScaleFactor = 0.6;
    if (o.sub) {
      const u = card.addText("до " + (tm.end ? tm.end.text : (cls.end_time || "")));
      u.font = Font.mediumSystemFont(11 * s);
      u.textColor = col(P.sub);
      u.lineLimit = 1;
    }
  } else {
    const big = card.addText(tm.start ? tm.start.text : (cls.start_time || ""));
    big.font = Font.boldRoundedSystemFont(o.big);
    big.textColor = col(P.text);
    big.lineLimit = 1;
    big.minimumScaleFactor = 0.6;
    if (o.sub && tm.startDate && tm.startDate > new Date()) {
      const rel = card.addDate(tm.startDate);
      rel.applyRelativeStyle();
      rel.font = Font.mediumSystemFont(11 * s);
      rel.textColor = col(P.sub);
      rel.lineLimit = 1;
    }
  }

  card.addSpacer(o.compact ? 1 : 3 * s);
  const subj = card.addText(cls.name_subject || "");
  subj.font = Font.semiboldSystemFont((o.compact ? 11.5 : 12.5) * s);
  subj.textColor = col(P.text);
  subj.lineLimit = o.compact ? 1 : 2;
  subj.minimumScaleFactor = 0.75;

  card.addSpacer();

  addEnds(card, tm.start ? tm.start.text : (cls.start_time || ""), tm.end ? tm.end.text : (cls.end_time || ""), s);
  card.addSpacer(o.compact ? 2 : 3);
  addFlexBar(card, now ? progressOf(tm) : 0, accent);
}

function addEmptyCard(parent, now, o) {
  const card = parent.addStack();
  card.layoutVertically();
  card.size = new Size(o.w, o.h);
  card.backgroundColor = col(P.pill, P.pillA);
  card.cornerRadius = o.radius;
  card.setPadding(o.pv, o.ph, o.pv, o.ph);

  const l = card.addText(now ? "СЕЙЧАС" : "ДАЛЬШЕ");
  l.font = Font.heavyRoundedSystemFont(10.5 * o.s);
  l.textColor = col(P.sub);
  l.lineLimit = 1;
  card.addSpacer();
  const t = card.addText(now ? "Пары нет" : "Пар больше нет");
  t.font = Font.boldRoundedSystemFont(15 * o.s);
  t.textColor = col(P.sub);
  t.lineLimit = 1;
  t.minimumScaleFactor = 0.7;
  card.addSpacer();
}

function addFooter(parent, data, size) {
  if (!data.updated_at) return;
  const f = parent.addStack();
  f.layoutHorizontally();
  f.addSpacer();
  const t = f.addText("Обновлено в " + data.updated_at);
  t.font = Font.systemFont(size);
  t.textColor = col(P.sub, 0.7);
}

function buildWide(widget, data) {
  const cur = data.current_class, nxt = data.next_class;
  const exam = data.has_exam && data.exam ? data.exam : null;
  const large = IS_LARGE;

  const padH = large ? 16 : 12, padT = large ? 16 : 10, padB = large ? 12 : 10;
  widget.setPadding(padT, padH, padB, padH);
  const innerW = WIDE_W - padH * 2;

  const bannerH = large ? 30 : 26, bannerGap = large ? 10 : 6;
  if (exam) {
    addExamBanner(widget, exam, large ? 12.5 : 11, innerW, bannerH);
    widget.addSpacer(bannerGap);
  }
  const examH = exam ? bannerH + bannerGap : 0;

  if (!cur && !nxt) {
    widget.addSpacer();
    const t = widget.addText("На сегодня всё");
    t.font = Font.boldRoundedSystemFont(large ? 30 : 24);
    t.textColor = col(P.text);
    widget.addSpacer(2);
    const s = widget.addText("Пар больше нет");
    s.font = Font.systemFont(large ? 15 : 13);
    s.textColor = col(P.sub);
    widget.addSpacer();
    if (large) addFooter(widget, data, 10);
    return;
  }

  if (!large) {
    const gap = 8;
    const cw = Math.floor((innerW - gap) / 2);
    const ch = MED_H - padT - padB - examH;
    const compact = Boolean(exam);
    const o = { w: cw, h: ch, s: 1, pv: compact ? 6 : 8, ph: 10, radius: 14, big: compact ? 24 : 28, compact: compact, sub: false };
    const row = widget.addStack();
    row.layoutHorizontally();
    if (cur) addCard(row, cur, true, o); else addEmptyCard(row, true, o);
    row.addSpacer(gap);
    if (nxt) addCard(row, nxt, false, o); else addEmptyCard(row, false, o);
    return;
  }

  const footerH = 16, gap = 10;
  const ch = Math.floor((LRG_H - padT - padB - examH - footerH - gap) / 2);
  const roomy = ch >= 130;
  const o = { w: innerW, h: ch, s: 1.1, pv: 12, ph: 14, radius: 18, big: roomy ? 40 : 34, compact: false, sub: roomy };
  if (cur) addCard(widget, cur, true, o); else addEmptyCard(widget, true, o);
  widget.addSpacer(gap);
  if (nxt) addCard(widget, nxt, false, o); else addEmptyCard(widget, false, o);
  widget.addSpacer();
  addFooter(widget, data, 10);
}

function buildAccessory(widget, data) {
  const cur = data.current_class, nxt = data.next_class;
  const cls = cur || nxt;

  if (FAMILY === "accessoryCircular") {
    widget.addAccessoryWidgetBackground = true;
    if (!cls) {
      const t = widget.addText("Всё");
      t.font = Font.boldRoundedSystemFont(16);
      return;
    }
    const tm = lessonTimes(cls);
    const a = widget.addText(room(cls.audience) || "–");
    a.font = Font.boldRoundedSystemFont(20);
    a.centerAlignText();
    const b = widget.addText(cur ? "до " + (tm.end ? tm.end.text : "") : (tm.start ? tm.start.text : ""));
    b.font = Font.systemFont(10);
    b.centerAlignText();
    return;
  }

  if (FAMILY === "accessoryInline") {
    if (!cls) { widget.addText("Пар больше нет"); return; }
    const tm = lessonTimes(cls);
    const when = cur ? "до " + (tm.end ? tm.end.text : "") : (tm.start ? tm.start.text : "");
    widget.addText(when + ", ауд. " + (room(cls.audience) || "–") + ", " + (cls.name_subject || ""));
    return;
  }

  if (!cls) {
    const t = widget.addText("Пар больше нет");
    t.font = Font.semiboldSystemFont(13);
    return;
  }
  const tm = lessonTimes(cls);
  const h = widget.addText((cur ? "Сейчас, до " + (tm.end ? tm.end.text : "") : "Дальше, " + tm.range));
  h.font = Font.semiboldSystemFont(11);
  h.lineLimit = 1;
  const s = widget.addText(cls.name_subject || "");
  s.font = Font.boldSystemFont(13);
  s.lineLimit = 2;
  s.minimumScaleFactor = 0.8;
  const r = widget.addText("ауд. " + (room(cls.audience) || "–"));
  r.font = Font.systemFont(11);
  r.lineLimit = 1;
}

const ERRORS = {
  auth: ["Неверный ключ", "Получите новый скрипт: откройте бота и отправьте /widget"],
  journal: ["Журнал недоступен", "Сайт журнала не отвечает. Виджет обновится сам"],
  net: ["Нет связи", "Проверьте интернет. Нажмите на виджет, чтобы обновить"],
  server: ["Ошибка сервера", "Нажмите на виджет, чтобы обновить"]
};

function buildError(widget, code) {
  const e = ERRORS[code] || ERRORS.server;
  if (IS_ACCESSORY) {
    const t = widget.addText(e[0]);
    t.font = Font.semiboldSystemFont(12);
    if (FAMILY === "accessoryRectangular") {
      const s = widget.addText(e[1]);
      s.font = Font.systemFont(10);
      s.minimumScaleFactor = 0.7;
    }
    return;
  }
  widget.setPadding(14, 16, 14, 16);
  const t = widget.addText(e[0]);
  t.font = Font.boldRoundedSystemFont(IS_SMALL ? 16 : 20);
  t.textColor = col(P.exam);
  widget.addSpacer(4);
  const s = widget.addText(e[1]);
  s.font = Font.systemFont(IS_SMALL ? 11 : 13);
  s.textColor = col(P.sub);
  s.minimumScaleFactor = 0.8;
}

let result = await loadData();
let widget = new ListWidget();

let tint = null;
if (result.data) {
  tint = (result.data.has_exam && result.data.exam) ? P.exam : (result.data.current_class ? P.now : (result.data.next_class ? P.next : null));
}
if (!IS_ACCESSORY) setBackground(widget, tint);
widget.url = "scriptable:///run/" + encodeURIComponent(Script.name()) + "?force=1";
// Это лишь пожелание: когда обновлять виджет, решает iOS
widget.refreshAfterDate = new Date(Date.now() + 10 * 60 * 1000);

if (!result.data) {
  buildError(widget, result.error);
} else if (IS_ACCESSORY) {
  buildAccessory(widget, result.data);
} else if (IS_SMALL) {
  buildSmall(widget, result.data);
} else {
  buildWide(widget, result.data);
}

if (config.runsInWidget) {
  Script.setWidget(widget);
} else if (IS_SMALL) {
  await widget.presentSmall();
} else if (IS_LARGE) {
  await widget.presentLarge();
} else {
  await widget.presentMedium();
}
Script.complete();
