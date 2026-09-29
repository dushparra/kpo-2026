package studying.exception;

/** Codes that identify errors raised by the report application. */
public enum ApplicationErrorCode {
    /** Input data failed validation. */
    VALIDATION_ERROR,
    /** A report could not be written to storage. */
    FILE_WRITE_ERROR,
    /** A requested service could not be found. */
    SERVICE_NOT_FOUND
}
