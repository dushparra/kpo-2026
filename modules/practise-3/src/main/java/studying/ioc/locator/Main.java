package studying.ioc.locator;

import java.time.LocalDateTime;
import studying.model.Report;
import studying.service.ReportSaver;
import studying.service.ReportSender;
import studying.service.impl.ReportSaverImpl;
import studying.service.impl.ReportSenderImpl;

/** Runs the Service Locator example. */
public final class Main {
    /** Number of cars in the sample report. */
    private static final int CARS_SOLD = 100;
    /** Number of motorcycles in the sample report. */
    private static final int MOTORCYCLES_SOLD = 50;

    private Main() {
    }

    /**
     * Registers dependencies and processes a sample report.
     *
     * @param arguments command-line arguments
     */
    public static void main(final String[] arguments) {
        var serviceLocator = new ServiceLocator();
        serviceLocator.register(ReportSaver.class, new ReportSaverImpl());
        serviceLocator.register(ReportSender.class, new ReportSenderImpl());

        var reportService = new ReportService(serviceLocator);
        reportService.process(createReport(), "example@example.com");
    }

    private static Report createReport() {
        var now = LocalDateTime.now();
        return Report.builder()
                .title("Отчёт Service Locator")
                .date(now.toLocalDate())
                .time(now.toLocalTime())
                .carsSold(CARS_SOLD)
                .motorcyclesSold(MOTORCYCLES_SOLD)
                .build();
    }
}
