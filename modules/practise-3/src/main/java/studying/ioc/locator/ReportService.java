package studying.ioc.locator;

import studying.exception.ApplicationErrorCode;
import studying.exception.ApplicationException;
import studying.model.Report;
import studying.service.ReportSaver;
import studying.service.ReportSender;

/** Processes reports by resolving dependencies through a service locator. */
public final class ReportService {
    /** Container used to find service implementations. */
    private final ServiceLocator serviceLocator;

    /**
     * Creates a report service backed by the supplied locator.
     *
     * @param locator service container
     */
    public ReportService(final ServiceLocator locator) {
        if (locator == null) {
            throw new ApplicationException(
                    ApplicationErrorCode.VALIDATION_ERROR,
                    "Service Locator должен быть указан"
            );
        }
        this.serviceLocator = locator;
    }

    /**
     * Saves a report and then sends it to a recipient.
     *
     * @param report report to process
     * @param email recipient email address
     */
    public void process(final Report report, final String email) {
        validate(report, email);

        var reportSaver = serviceLocator.resolve(ReportSaver.class);
        var reportSender = serviceLocator.resolve(ReportSender.class);

        reportSaver.save(report);
        reportSender.send(report, email);
    }

    private static void validate(
            final Report report,
            final String email
    ) {
        if (report == null || email == null || email.isBlank()) {
            throw new ApplicationException(
                    ApplicationErrorCode.VALIDATION_ERROR,
                    "Отчёт должен существовать, а email не должен быть пустым"
            );
        }
    }
}
