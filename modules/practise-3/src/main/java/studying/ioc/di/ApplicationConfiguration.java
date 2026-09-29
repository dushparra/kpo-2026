package studying.ioc.di;

import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import studying.service.ReportSaver;
import studying.service.ReportSender;
import studying.service.impl.ReportSaverImpl;
import studying.service.impl.ReportSenderImpl;

/** Describes the object graph assembled by the Spring IoC container. */
@Configuration(proxyBeanMethods = false)
public final class ApplicationConfiguration {
    /**
     * Creates the report persistence component.
     *
     * @return report saver
     */
    @Bean
    public ReportSaver reportSaver() {
        return new ReportSaverImpl();
    }

    /**
     * Creates the report delivery component.
     *
     * @return report sender
     */
    @Bean
    public ReportSender reportSender() {
        return new ReportSenderImpl();
    }

    /**
     * Creates a report service with dependencies supplied by Spring.
     *
     * @param saver report persistence component
     * @param sender report delivery component
     * @return configured report service
     */
    @Bean
    public ReportService reportService(
            final ReportSaver saver,
            final ReportSender sender
    ) {
        return new ReportService(saver, sender);
    }
}
