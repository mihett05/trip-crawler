workspace "Travel Crawler" "Сервис для составления комплексный маршрутов путешествий по России" {

    !identifiers hierarchical

    model {
        archetypes {
            service = container {
                technology "Go"
                tags "Application"
            }
            database = container {
                tags "Database"
            }
            queue = container {
                tags "Queue"
            }
        }


        user = person "Путешественник" "Человек, ищущий маршрут путешествия по запросу"


        s = softwareSystem "Наш сервис" {
            graphDbms = database "Графовая БД" "Для хранения информации о маршрутах" "Dgraph"
            pointsDbms = database "Кэш точек" "Денормализация конечных/начальных точек для быстрой выдачи" "Valkey"

            queue = queue "Очередь" "Брокер для получения задач на парсинг" "NATS"

            parser = service "Парсер" "Сервис для получения структуры маршрутов и доступности" {
                transportAdapter = component "TransportAdapter" "Адаптер для получения маршрутов и доступности из конкретного источника"
                queueHandler = component "QueueHandler" "Хендлер запросов на парсинг"
                queueHandler -> transportAdapter
            }
            queue -> parser.queueHandler
            
            mainService = service "Main Service" "Основной сервис" {
                graphRepository = component "GraphRepository" "Репозиторий для работы с данными внутри графа"
                pointsRepository = component "PointsRepository" "Репозиторий для поиска узлов"
                itineraryBuilder = component "ItineraryBuilder" "Реализация алгоритма построения маршрута"
                schedulerQueueAdapter = component "SchedulerQueueAdapter" "Адаптер над очередью планировщика"

                itineraryBuilderService = component "ItineraryBuilderService" "Сервис для построения маршрутов"
                graphWriterService = component "GraphWriterService" "Сервис для сохранения информации о частях графа"
                pointsSearchService = component "PointsSearchService" "Сервис для поиска точек маршрута"
                graphScheduler = component "GraphScheduler" "Планировщик задач на парсинг"

                itineraryBuilderService -> graphRepository
                itineraryBuilderService -> itineraryBuilder
                graphWriterService -> graphRepository
                pointsSearchService -> pointsRepository
                graphScheduler -> pointsRepository
                graphScheduler -> schedulerQueueAdapter
                
                itineraryHttpHandler = component "ItineraryHandler" "Хендлеры для работы с маршрутами"
                graphQueueHandler = component "GraphHandler" "Хендлер для сохранения информации об изменениях в графе"
                pointsHttpHandler = component "PointsHandler" "Хендлеры для работы с доступными точками маршрута"

                itineraryHttpHandler -> itineraryBuilderService
                graphQueueHandler -> graphWriterService
                pointsHttpHandler -> pointsSearchService
            }

            mainService.schedulerQueueAdapter -> queue
            mainService.pointsRepository -> pointsDbms
            mainService.graphRepository -> graphDbms
            mainService.itineraryBuilder -> graphDbms

            frontend = container "Frontend" "Фронтенд для поиска и бронирования" "React"

            

            frontend -> mainService "API" "REST"
        }

        rzhd = softwareSystem "РЖД" "Незадокументированное api для получения информации о ценах + csv таблица для связей между станциями"
        aviasales = softwareSystem "Aviasales" "Использование api для получения статических данных о связях аэропортов"
        airlines = softwareSystem "Сайты авиакомпаний" "Сайты или api авиакомпаний для цен и времени"

        user -> s.frontend "Ищет маршруты, бронирует билеты и отели"
        

        s.parser -> rzhd
        s.parser -> aviasales
        s.parser -> airlines
    }

    views {
        systemContext s "SystemContext" {
            include *
            autolayout lr
        }

        container s "Service" {
            include *
            autolayout lr
        }

        component s.mainService "MainService" {
            include *
            autolayout lr
        }

        component s.parser "Parser" {
            include *
            autolayout lr
        }

        styles {
            element "Element" {
                color #ffffff
            }
            element "Person" {
                background #048c04
                shape person
            }
            element "Software System" {
                background #047804
            }
            element "Container" {
                background #55aa55
            }
            element "Component" {
                background #55aa55
            }
            element "Database" {
                shape cylinder
            }
            element "Queue" {
                shape pipe
            }
        }
    }

    configuration {
        scope softwaresystem
    }
}