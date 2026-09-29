package studying.ioc.di;

import studying.exception.ApplicationErrorCode;
import studying.exception.ApplicationException;
import studying.model.Report;
import studying.service.ReportSaver;
import studying.service.ReportSender;

/** Processes reports using explicitly injected dependencies. */
public final class ReportService {
    /** Component responsible for saving reports. */
    private final ReportSaver reportSaver;
    /** Component responsible for sending reports. */
    private final ReportSender reportSender;

    /**
     * Creates a report service with all required dependencies.
     *
     * @param saver report persistence component
     * @param sender report delivery component
     */
    public ReportService(
            final ReportSaver saver,
            final ReportSender sender
    ) {
        if (saver == null || sender == null) {
            throw new ApplicationException(
                    ApplicationErrorCode.VALIDATION_ERROR,
                    "Зависимости ReportService должны быть указаны"
            );
        }
        this.reportSaver = saver;
        this.reportSender = sender;
    }

    /**
     * Saves a report and then sends it to a recipient.
     *
     * @param report report to process
     * @param email recipient email address
     */
    public void process(final Report report, final String email) {
        if (report == null || email == null || email.isBlank()) {
            throw new ApplicationException(
                    ApplicationErrorCode.VALIDATION_ERROR,
                    "Отчёт должен существовать, а email не должен быть пустым"
            );
        }

        reportSaver.save(report);
        reportSender.send(report, email);
    }
}
