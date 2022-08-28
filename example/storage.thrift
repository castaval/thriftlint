include "base.thrift"

namespace java org.thrift.cds.storage
namespace erlang cds

/** Дата экспирации */
struct ExpDate {
    /** Месяц 1..12 */
    1: required i8 month
    /** Год 2015..∞ */
    2: required i16 year
}

/** Карточные данные */
/** NOTE: Код верификации хранится, при необходимости, в данных сессии */
struct CardData {
    /** Номер карты без пробелов [0-9]{14,19} */
    1: required string pan
    2: optional ExpDate exp_date
    3: optional string cardholder_name
}

struct PutCardData {
    /** Номер карты без пробелов [0-9]{14,19} */
    1: required string pan
}

struct PutCardResult {
    1: required base.BankCard bank_card
}

/** Код проверки подлинности банковской карты */
struct CardSecurityCode {
    /** Код верификации [0-9]{3,4} */
    1: required string value
}

/** Данные, необходимые для авторизации по 3DS протоколу */
struct Auth3DS {
    /** Криптограмма для проверки подлинности */
    1: required string cryptogram
    /** Тип транзакции */
    2: optional string eci
}

/** Данные, необходимые для проверки подлинности банковской карты */
union AuthData {
    1: CardSecurityCode card_security_code
    2: Auth3DS auth_3ds
}

/** Данные сессии */
struct SessionData {
    1: required AuthData auth_data
}

exception InvalidCardData {
    1: optional string reason
}

exception CardDataNotFound {}

exception SessionDataNotFound {}

/**
 * Интерфейс для приложений
 *
 * При недоступности (отсутствии или залоченности) кейринга сервис сигнализирует об этом с помощью
 * системной ошибки `Resource Unavailable`.
 */
service Storage {

    /** Получить карточные данные */
    CardData GetCardData (1: base.Token token)
        throws (1: CardDataNotFound not_found)

    /** Получить данные сессии */
    SessionData GetSessionData (1: base.PaymentSessionID session_id)
        throws (1: SessionDataNotFound not_found)

    /** Сохранить карточные данные */
    PutCardResult PutCard (1: PutCardData card_data)
        throws (1: InvalidCardData invalid)

    /** Сохранить сессионные данные */
    void PutSession (1: base.PaymentSessionID session_id, 2: SessionData session_data)

}