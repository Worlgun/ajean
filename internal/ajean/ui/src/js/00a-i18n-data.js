// 00a-i18n-data.js — dictionnaires de traduction. Un objet par langue, même
// jeu de clés partout (le français est le texte source, référence en cas de
// trou ailleurs — voir t() dans 00b-i18n.js). Nommage des clés :
// "section.sous-clé" (namespacé par panneau pour éviter les collisions au fil
// des extractions).
//
// Pour ajouter une langue : copier un bloc entier ci-dessous, traduire les
// valeurs (JAMAIS les clés), ajouter le code dans LANG_NAMES (00b-i18n.js) et
// dans #lang-select (index.tmpl.html). Rien d'autre à toucher.
var I18N = {
fr: {
  "appearance.title": "Apparence",
  "appearance.dark_mode": "mode sombre",
  "appearance.language": "Langue",
  "appearance.display_subhead": "Affichage",
  "appearance.hide_reasoning": "masquer le raisonnement",
  "appearance.hide_tools": "masquer les appels d'outils",
  "appearance.fold_tools": "garder les bulles repliées",
  "appearance.fold_tools_help": "Les bulles de raisonnement et d'outil arrivent <b>déjà fermées</b> au lieu de s'ouvrir puis se refermer. Un clic sur leur étiquette les déplie.",
  "appearance.hide_side": "barre latérale escamotable",
  "appearance.hide_side_help": "La barre latérale devient un tiroir <b>même sur ordinateur</b> : elle reste fermée et s'ouvre avec le bouton ☰. Toute la largeur pour la conversation.",
  "appearance.enter_newline": "Entrée = retour à la ligne",
  "appearance.enter_newline_help": "Par défaut, <b>Entrée</b> envoie le message et <b>Maj+Entrée</b> fait un retour à la ligne. Activez ceci pour inverser : Entrée va à la ligne, et <b>Maj+Entrée</b> (ou Ctrl+Entrée) envoie."
},
en: {
  "appearance.title": "Appearance",
  "appearance.dark_mode": "dark mode",
  "appearance.language": "Language",
  "appearance.display_subhead": "Display",
  "appearance.hide_reasoning": "hide reasoning",
  "appearance.hide_tools": "hide tool calls",
  "appearance.fold_tools": "keep bubbles collapsed",
  "appearance.fold_tools_help": "Reasoning and tool bubbles now arrive <b>already collapsed</b> instead of opening then closing. Click their label to expand them.",
  "appearance.hide_side": "collapsible sidebar",
  "appearance.hide_side_help": "The sidebar becomes a drawer <b>even on desktop</b>: it stays closed and opens with the ☰ button. Full width for the conversation.",
  "appearance.enter_newline": "Enter = new line",
  "appearance.enter_newline_help": "By default, <b>Enter</b> sends the message and <b>Shift+Enter</b> makes a new line. Turn this on to flip it: Enter makes a new line, and <b>Shift+Enter</b> (or Ctrl+Enter) sends."
},
it: {
  "appearance.title": "Aspetto",
  "appearance.dark_mode": "modalità scura",
  "appearance.language": "Lingua",
  "appearance.display_subhead": "Visualizzazione",
  "appearance.hide_reasoning": "nascondi il ragionamento",
  "appearance.hide_tools": "nascondi le chiamate agli strumenti",
  "appearance.fold_tools": "mantieni le bolle ripiegate",
  "appearance.fold_tools_help": "Le bolle di ragionamento e strumento ora arrivano <b>già ripiegate</b> invece di aprirsi e richiudersi. Un clic sull'etichetta le espande.",
  "appearance.hide_side": "barra laterale a scomparsa",
  "appearance.hide_side_help": "La barra laterale diventa un cassetto <b>anche su desktop</b>: resta chiusa e si apre con il pulsante ☰. Larghezza piena per la conversazione.",
  "appearance.enter_newline": "Invio = a capo",
  "appearance.enter_newline_help": "Per impostazione predefinita, <b>Invio</b> invia il messaggio e <b>Maiusc+Invio</b> va a capo. Attiva questa opzione per invertire: Invio va a capo, e <b>Maiusc+Invio</b> (o Ctrl+Invio) invia."
},
es: {
  "appearance.title": "Apariencia",
  "appearance.dark_mode": "modo oscuro",
  "appearance.language": "Idioma",
  "appearance.display_subhead": "Visualización",
  "appearance.hide_reasoning": "ocultar el razonamiento",
  "appearance.hide_tools": "ocultar las llamadas a herramientas",
  "appearance.fold_tools": "mantener las burbujas plegadas",
  "appearance.fold_tools_help": "Las burbujas de razonamiento y herramientas ahora llegan <b>ya plegadas</b> en lugar de abrirse y luego cerrarse. Un clic en su etiqueta las despliega.",
  "appearance.hide_side": "barra lateral retráctil",
  "appearance.hide_side_help": "La barra lateral se convierte en un cajón <b>incluso en escritorio</b>: permanece cerrada y se abre con el botón ☰. Todo el ancho para la conversación.",
  "appearance.enter_newline": "Intro = salto de línea",
  "appearance.enter_newline_help": "Por defecto, <b>Intro</b> envía el mensaje y <b>Mayús+Intro</b> hace un salto de línea. Activa esto para invertirlo: Intro salta de línea, y <b>Mayús+Intro</b> (o Ctrl+Intro) envía."
},
ru: {
  "appearance.title": "Внешний вид",
  "appearance.dark_mode": "тёмная тема",
  "appearance.language": "Язык",
  "appearance.display_subhead": "Отображение",
  "appearance.hide_reasoning": "скрыть рассуждения",
  "appearance.hide_tools": "скрыть вызовы инструментов",
  "appearance.fold_tools": "держать пузыри свёрнутыми",
  "appearance.fold_tools_help": "Пузыри рассуждений и инструментов теперь приходят <b>уже свёрнутыми</b>, вместо того чтобы открываться и закрываться. Клик по ярлыку разворачивает их.",
  "appearance.hide_side": "сворачиваемая боковая панель",
  "appearance.hide_side_help": "Боковая панель становится выдвижной <b>даже на компьютере</b>: она остаётся закрытой и открывается кнопкой ☰. Вся ширина отдана беседе.",
  "appearance.enter_newline": "Enter = новая строка",
  "appearance.enter_newline_help": "По умолчанию <b>Enter</b> отправляет сообщение, а <b>Shift+Enter</b> делает новую строку. Включите это, чтобы поменять местами: Enter — новая строка, а <b>Shift+Enter</b> (или Ctrl+Enter) отправляет."
},
de: {
  "appearance.title": "Erscheinungsbild",
  "appearance.dark_mode": "dunkler Modus",
  "appearance.language": "Sprache",
  "appearance.display_subhead": "Anzeige",
  "appearance.hide_reasoning": "Überlegungen ausblenden",
  "appearance.hide_tools": "Tool-Aufrufe ausblenden",
  "appearance.fold_tools": "Blasen eingeklappt halten",
  "appearance.fold_tools_help": "Überlegungs- und Tool-Blasen kommen jetzt <b>bereits eingeklappt</b> an, statt sich zu öffnen und wieder zu schließen. Ein Klick auf ihre Beschriftung klappt sie auf.",
  "appearance.hide_side": "einklappbare Seitenleiste",
  "appearance.hide_side_help": "Die Seitenleiste wird <b>auch am Desktop</b> zu einer Schublade: Sie bleibt geschlossen und öffnet sich über die ☰-Schaltfläche. Volle Breite für die Unterhaltung.",
  "appearance.enter_newline": "Enter = neue Zeile",
  "appearance.enter_newline_help": "Standardmäßig sendet <b>Enter</b> die Nachricht und <b>Umschalt+Enter</b> erzeugt eine neue Zeile. Aktiviere dies, um es umzukehren: Enter erzeugt eine neue Zeile, und <b>Umschalt+Enter</b> (oder Strg+Enter) sendet."
}
};
