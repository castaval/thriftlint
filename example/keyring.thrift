namespace java org.thrift.cds.keyring
namespace erlang cds

typedef string ShareholderId;

typedef i64 KeyId;

/** Подписанная часть мастер ключа и ID подписавшего */
struct SignedMasterKeyShare {
    1: required ShareholderId id
    2: required binary signed_share
}

/** Зашифрованная часть мастер-ключа и кому он предназначается */
struct EncryptedMasterKeyShare {
    // Уникальный ID, для однозначного определения владения
    1: required ShareholderId id
    // Неуникальный идентификатор с ФИО/email/etc владельца
    2: required string owner
    // Зашифрованный MasterKeyShare
    3: required binary encrypted_share
}

typedef list<EncryptedMasterKeyShare> EncryptedMasterKeyShares;

struct Success {}

union KeyringOperationStatus {
    /** Успешно. */
    1: Success success
    /** Сколько частей ключа нужно еще ввести, чтобы провести манипуляцию над Keyring. */
    2: i16 more_keys_needed
}

enum Initialization {
    uninitialized
    validation
}

enum Rekeying {
    uninitialized
    confirmation
    postconfirmation
    validation
}

enum Rotation {
    uninitialized
    validation
}

enum Unlock {
    uninitialized
    validation
}

enum Status {
    not_initialized
    unlocked
    locked
}

union Activity {
    1: Initialization initialization
    2: Rekeying rekeying
    3: Rotation rotation
    4: Unlock unlock
}

typedef list<Activity> Activities;

typedef i16 ShareId

typedef map<ShareId, ShareholderId> ShareSubmitters;

typedef i32 Seconds;

struct RotationState {
    1: required Rotation phase
    2: optional Seconds lifetime
    3: required ShareSubmitters confirmation_shares
}

struct InitializationState {
    1: required Initialization phase
    2: optional Seconds lifetime
    3: required ShareSubmitters validation_shares
}

struct UnlockState {
    1: required Unlock phase
    2: optional Seconds lifetime
    3: required ShareSubmitters confirmation_shares
}

struct RekeyingState {
    1: required Rekeying phase
    2: optional Seconds lifetime
    3: required ShareSubmitters confirmation_shares
    4: required ShareSubmitters validation_shares
}

struct ActivitiesState {
    1: required InitializationState initialization
    2: required RotationState rotation
    3: required UnlockState unlock
    4: required RekeyingState rekeying
}

struct KeyringState {
    1: required Status status
    2: required ActivitiesState activities
}

struct KeyMeta {
    1: required bool retired
    2: required SecurityParameters security_parameters
}

struct KeyMetaDiff {
    1: required bool retired
}

struct KeyringMeta {
    1: required map<KeyId, KeyMeta> keys_meta
    2: required KeyId current_key_id
}

struct KeyringMetaDiff {
    1: optional map<KeyId, KeyMetaDiff> keys_meta
    2: optional KeyId current_key_id
}

struct ScryptOptions {
    1: required i32 n
    2: required i32 r
    3: required i32 p
}

struct SecurityParameters {
    1: required ScryptOptions deduplication_hash_opts
}

exception InvalidStatus {
    1: required Status status
}

exception InvalidActivity {
    1: required Activity activity
}

exception InvalidArguments {
    1: optional string reason
}

exception InvalidKeyringMeta {
    1: optional string reason
}

exception OperationAborted {
    1: optional string reason
}

exception VerificationFailed {}

/** Интерфейс для администраторов */
service KeyringManagement {
    
    /** Создать новый кейринг при начальном состоянии
     *  threshold - минимально необходимое количество ключей для восстановления мастер ключа
     */
    EncryptedMasterKeyShares StartInit (1: i16 threshold)
        throws (1: InvalidStatus invalid_status,
                2: InvalidActivity invalid_activity,
                3: InvalidArguments invalid_args)

    /** Валидирует и завершает операцию над Keyring
     *  Вызывается после Init и Rekey (CDS-25)
     */
    KeyringOperationStatus ValidateInit (1: SignedMasterKeyShare key_share)
        throws (1: InvalidStatus invalid_status,
                2: InvalidActivity invalid_activity,
                3: VerificationFailed verification_failed,
                // Исключения ниже переводят машину в состояние `uninitialized`
                4: OperationAborted operation_aborted)

    /** Отменяет Init не прошедший валидацию и дает возможность запустить его заново */
    void CancelInit () throws (1: InvalidStatus invalid_status)

    /** Создать новый masterkey при наличии уже имеющегося
     *  threshold - минимально необходимое количество ключей для восстановления мастер ключа
     */
    void StartRekey (1: i16 threshold)
        throws (1: InvalidStatus invalid_status,
                2: InvalidActivity invalid_activity,
                3: InvalidArguments invalid_args)

    /** Подтвердить операцию создания нового masterkey
     *  key_share - старый masterkey share в количестве threshold
     */
    KeyringOperationStatus ConfirmRekey (1: SignedMasterKeyShare key_share)
        throws (1: InvalidStatus invalid_status,
                2: InvalidActivity invalid_activity,
                3: VerificationFailed verification_failed,
                4: OperationAborted operation_aborted)

    /** Начать валидацию операции и получить Зашифрованные masterkey share */
    EncryptedMasterKeyShares StartRekeyValidation ()
        throws (1: InvalidStatus invalid_status,
                2: InvalidActivity invalid_activity)

    /** Провалидировать расшифрованными фрагментами нового ключа
     *  key_share - новый masterkey share в количестве num
     */
    KeyringOperationStatus ValidateRekey (1: SignedMasterKeyShare key_share)
        throws (1: InvalidStatus invalid_status,
                2: InvalidActivity invalid_activity,
                3: VerificationFailed verification_failed,
                4: OperationAborted operation_aborted)

    /** Отменить операцию создания нового masterkey */
    void CancelRekey () throws (1: InvalidStatus invalid_status)

    /** Начинает процесс блокировки */
    void StartUnlock ()
        throws (1: InvalidStatus invalid_status,
                2: InvalidActivity invalid_activity)

    /** Предоставить часть мастер-ключа для расшифровки кейринга.
     *  Необходимо вызвать с разными частами мастер столько раз, сколько было указано в качестве
     *  параметра threshold при создании кейринга
     */
    KeyringOperationStatus ConfirmUnlock (1: SignedMasterKeyShare key_share)
        throws (1: InvalidStatus invalid_status,
                2: InvalidActivity invalid_activity,
                3: VerificationFailed verification_failed,
                4: OperationAborted operation_aborted)

    /** Отменяет процесс блокировки */
    void CancelUnlock () throws (1: InvalidStatus invalid_status)

    /** Зашифровать кейринг */
    void Lock () throws (1: InvalidStatus invalid_status)

    /** Начать процесс добавления нового ключа в кейринг */
    void StartRotate ()
        throws (1: InvalidStatus invalid_status,
                2: InvalidActivity invalid_activity)

    /*  Предоставить часть мастер-ключа для зашифровки нового инстанса кейринга.
     *  См. `Unlock`
     */
    KeyringOperationStatus ConfirmRotate (1: SignedMasterKeyShare key_share)
        throws (1: InvalidStatus invalid_status,
                2: InvalidActivity invalid_activity,
                3: VerificationFailed verification_failed,
                4: OperationAborted operation_aborted)

    /** Отменяет процесс добавления нового ключа в кейринг */
    void CancelRotate () throws (1: InvalidStatus invalid_status)

    /** Получить состояние операций */
    KeyringState GetState ()

    /** Дополнить метаданные Keyring, используемые Storage */
    void UpdateKeyringMeta (1: KeyringMetaDiff keyring_meta)
        throws (1: InvalidKeyringMeta invalid_meta
                2: InvalidStatus invalid_status)

    /** Получить текущие мета данные Keyring, используемые Storage */
    KeyringMeta GetKeyringMeta ()
}

typedef binary KeyData;

struct Key {
    1: required KeyData data
    2: required KeyMeta meta
}

struct Keyring {
    1: required i64 version
    2: required KeyId current_key_id
    3: required map<KeyId, Key> keys
}

/** Интерфейс для получения ключей */
service KeyringStorage {

    /** Возвращает все ключи c метаданными, а также идентификатор текущего ключа */
    Keyring GetKeyring ()
        throws (1: InvalidStatus invalid_status)

}
