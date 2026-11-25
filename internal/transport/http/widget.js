// Ключ пользователя. Бот подставляет сюда персональный ключ вместо {{KEY}}
const KEY = "{{KEY}}";

// Адрес сервера (новый сервер работает на порту 8081)
// Если скрипт запущен нажатием на виджет (?force=1), просим сервер обновить данные
const FORCE = (args.queryParameters && args.queryParameters.force === "1") ? "&force=1" : "";
const API_URL = "http://widget.sqardey.xyz/api/widget?key=" + encodeURIComponent(KEY) + FORCE;

// Загружаем данные и возвращаем { data, error }
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

let result = await loadData();
let data = result.data;

let widget = new ListWidget();
widget.backgroundColor = new Color("#161618");
widget.setPadding(12, 14, 12, 14);

// Нажатие на виджет запускает этот же скрипт с force=1 (принудительное обновление)
widget.url = "scriptable:///run/" + encodeURIComponent(Script.name()) + "?force=1";

if (!data) {
  let err = widget.addText("⚠️ " + result.error);
  err.textColor = new Color("#FF453A");
  err.font = Font.boldSystemFont(13);
  let sub = widget.addText("Проверь ключ, сервер или сеть");
  sub.textColor = new Color("#8E8E93");
  sub.font = Font.systemFont(11);
} else {
  let hasExam = Boolean(data.has_exam);
  let examObj = data.exam;

  // 1. Экзамен (всегда сверху)
  if (hasExam && examObj) {
    let examName = examObj.name_subject || "Экзамен";

    let exBadge = widget.addText("🔥 ЭКЗАМЕН СЕГОДНЯ");
    exBadge.textColor = new Color("#FF453A");
    exBadge.font = Font.boldSystemFont(11);

    let exSubj = widget.addText(examName);
    exSubj.textColor = Color.white();
    exSubj.font = Font.boldSystemFont(13);
    exSubj.lineLimit = 1;
    widget.addSpacer(6);
  }

  // 2. Текущая пара
  let cur = data.current_class;
  if (cur) {
    let badge = widget.addText(`ИДЁТ СЕЙЧАС (до ${cur.end_time})`);
    badge.textColor = new Color("#30D158");
    badge.font = Font.boldSystemFont(10);

    let name = widget.addText(`${cur.name_subject} • ${cur.audience} ауд.`);
    name.textColor = Color.white();
    name.font = Font.semiboldSystemFont(13);
    name.lineLimit = 1;
    widget.addSpacer(6);
  }

  // 3. Следующая пара
  let nextHeader = widget.addText("СЛЕДУЮЩАЯ ПАРА");
  nextHeader.textColor = new Color("#8E8E93");
  nextHeader.font = Font.boldSystemFont(10);
  widget.addSpacer(2);

  let nxt = data.next_class;
  if (nxt) {
    let timeText = widget.addText(`${nxt.start_time} – ${nxt.end_time} • ${nxt.audience} ауд.`);
    timeText.textColor = new Color("#0A84FF");
    timeText.font = Font.semiboldSystemFont(12);

    let subjText = widget.addText(nxt.name_subject);
    subjText.textColor = Color.white();
    subjText.font = Font.boldSystemFont(14);
    subjText.lineLimit = 2;
  } else {
    let free = widget.addText("Пар больше нет 🎉");
    free.textColor = new Color("#8E8E93");
    free.font = Font.mediumSystemFont(13);
  }

  widget.addSpacer();

  let footer = widget.addText(`Обновлено в ${data.updated_at || ""}`);
  footer.textColor = new Color("#48484A");
  footer.font = Font.systemFont(9);
  footer.rightAlignText();
}

if (config.runsInWidget) {
  Script.setWidget(widget);
} else {
  widget.presentMedium();
}
Script.complete();
