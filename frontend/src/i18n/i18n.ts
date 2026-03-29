import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';

// Переводы на английский
const en = {
  translation: {
    // Заголовки
    appTitle: 'Trip Crawler',
    appSubtitle: 'Discover amazing destinations and create unforgettable memories',
    routeDetails: 'Route Details',
    tripRouteMap: 'Trip Route Map',

    // Поля формы
    from: 'From',
    to: 'To',
    addStopoverCity: 'Add Stopover City',
    departureDate: 'Departure Date',
    stopoverCities: 'Stopover Cities (up to 3)',

    // Кнопки
    add: 'Add',
    planAnotherTrip: 'Plan Another Trip',
    findTrips: 'Find Trips',
    planning: 'Planning...',

    // Специальные метки
    start: 'START',
    end: 'END',

    // Тексты ошибок
    startingCityRequired: 'Starting city is required',
    destinationCityRequired: 'Destination city is required',
    startDateRequired: 'Start date is required',

    // Вспомогательные тексты
    citiesAdded: '{{count}}/3 cities added',
    days: '{{count}} day{{plural}}',

    // Результаты поездки
    yourTripPlan: 'Your Trip Plan',
    tripFromTo: 'Trip from {{origin}} to {{destination}}',
    journeyInfo: '{{days}} days journey • {{distance}} km',
    destinationsCount: '{{count}} Destinations',
    includingStops: 'Including stops',
    arrival: 'Arrival:',
    departure: 'Departure:',
    stay: 'Stay: {{count}} day{{plural}}',

    // Карта
    arrive: 'Arrive',
    depart: 'Depart',

    // Прочее
    monday: 'Mon',
    tuesday: 'Tue',
    wednesday: 'Wed',
    thursday: 'Thu',
    friday: 'Fri',
    saturday: 'Sat',
    sunday: 'Sun',
    january: 'Jan',
    february: 'Feb',
    march: 'Mar',
    april: 'Apr',
    may: 'May',
    june: 'Jun',
    july: 'Jul',
    august: 'Aug',
    september: 'Sep',
    october: 'Oct',
    november: 'Nov',
    december: 'Dec',

    // Transport types
    train: 'Train',
    bus: 'Bus',
    airplane: 'Airplane',
    transportType: 'Transport Type',
    availableTickets: 'Available Tickets',
    price: 'Price',

    // Min/Max Days
    minDays: 'Min Days',
    maxDays: 'Max Days',
    minMaxError: 'Max days must be greater than or equal to min days',
  },
};

// Переводы на русский
const ru = {
  translation: {
    // Заголовки
    appTitle: 'Поиск поездок',
    appSubtitle:
      'Откройте для себя удивительные достопримечательности и создайте незабываемые воспоминания',
    routeDetails: 'Детали маршрута',
    tripRouteMap: 'Карта маршрута поездки',

    // Поля формы
    from: 'Откуда',
    to: 'Куда',
    addStopoverCity: 'Добавить промежуточный город',
    departureDate: 'Дата отправления',
    stopoverCities: 'Промежуточные города (до 3)',

    // Кнопки
    add: 'Добавить',
    planAnotherTrip: 'Спланировать другую поездку',
    findTrips: 'Найти поездки',
    planning: 'Планирование...',

    // Специальные метки
    start: 'НАЧАЛО',
    end: 'КОНЕЦ',

    // Тексты ошибок
    startingCityRequired: 'Необходимо указать город начала',
    destinationCityRequired: 'Необходимо указать город назначения',
    startDateRequired: 'Необходимо указать дату начала',

    // Вспомогательные тексты
    citiesAdded: '{{count}}/3 городов добавлено',
    days: '{{count}} день{{plural}}',

    // Результаты поездки
    yourTripPlan: 'Ваш план поездки',
    tripFromTo: 'Поездка из {{origin}} в {{destination}}',
    journeyInfo: '{{days}} дней пути • {{distance}} км',
    destinationsCount: '{{count}} пунктов назначения',
    includingStops: 'включая остановки',
    arrival: 'Прибытие:',
    departure: 'Отправление:',
    stay: 'Пребывание: {{count}} день{{plural}}',

    // Карта
    arrive: 'Прибытие',
    depart: 'Отправление',

    // Прочее
    monday: 'Пн',
    tuesday: 'Вт',
    wednesday: 'Ср',
    thursday: 'Чт',
    friday: 'Пт',
    saturday: 'Сб',
    sunday: 'Вс',
    january: 'Янв',
    february: 'Фев',
    march: 'Мар',
    april: 'Апр',
    may: 'Май',
    june: 'Июн',
    july: 'Июл',
    august: 'Авг',
    september: 'Сен',
    october: 'Окт',
    november: 'Ноя',
    december: 'Дек',

    // Типы транспорта
    train: 'Поезд',
    bus: 'Автобус',
    airplane: 'Самолёт',
    transportType: 'Тип транспорта',
    availableTickets: 'Доступных билетов',
    price: 'Цена',

    // Мин/Макс дни
    minDays: 'Мин дни',
    maxDays: 'Макс дни',
    minMaxError: 'Макс дни должны быть больше или равны мин дням',
  },
};

i18n.use(initReactI18next).init({
  resources: {
    en: { ...en },
    ru: { ...ru },
  },
  lng: 'en', // Язык по умолчанию
  fallbackLng: 'en',
  interpolation: {
    escapeValue: false, // React уже защищает от XSS
  },
});

export default i18n;
