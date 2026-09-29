package studying.ioc.di;

import java.time.LocalDateTime;
import org.springframework.context.annotation.AnnotationConfigApplicationContext;
import studying.model.Report;

/** Runs the Spring Dependency Injection example. */
public final class Main {
    /** Number of cars in the sample report. */
    private static final int CARS_SOLD = 100;
    /** Number of motorcycles in the sample report. */
    private static final int MOTORCYCLES_SOLD = 50;

    private Main() {
    }

    /**
     * Starts Spring and processes a sample report.
     *
     * @param arguments command-line arguments
     */
    public static void main(final String[] arguments) {
        var context = new AnnotationConfigApplicationContext(
                ApplicationConfiguration.class
        );
        try (context) {
            var reportService = context.getBean(ReportService.class);
            reportService.process(createReport(), "example@example.com");
        }
    }

    private static Report createReport() {
        var now = LocalDateTime.now();
        return Report.builder()
                .title("Отчёт Dependency Injection")
                .date(now.toLocalDate())
                .time(now.toLocalTime())
                .carsSold(CARS_SOLD)
                .motorcyclesSold(MOTORCYCLES_SOLD)
                .build();
    }
}
