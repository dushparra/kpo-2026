package studying.service.impl;

import studying.exception.ApplicationErrorCode;
import studying.exception.ApplicationException;
import studying.model.Report;
import studying.service.ReportSender;

public final class ReportSenderImpl implements ReportSender {
    @Override
    public void send(final Report report, final String email) {
        if (report == null || email == null || email.isBlank()) {
            throw new ApplicationException(
                    ApplicationErrorCode.VALIDATION_ERROR,
                    "Отчёт должен существовать, а email не должен быть пустым"
            );
        }

        System.out.printf(
                "Отправка отчёта «%s» на email: %s%n",
                report.title(),
                email
        );
    }
}
