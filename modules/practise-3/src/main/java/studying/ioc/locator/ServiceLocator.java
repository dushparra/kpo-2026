package studying.ioc.locator;

import java.util.HashMap;
import java.util.Map;
import studying.exception.ApplicationErrorCode;
import studying.exception.ApplicationException;

/** Stores and resolves service implementations by their contracts. */
public final class ServiceLocator {
    /** Registered service implementations. */
    private final Map<Class<?>, Object> services = new HashMap<>();

    /**
     * Registers or replaces an implementation for a contract.
     *
     * @param <T> service contract type
     * @param contract service contract
     * @param implementation contract implementation
     */
    public <T> void register(
            final Class<T> contract,
            final T implementation
    ) {
        if (contract == null || implementation == null) {
            throw new ApplicationException(
                    ApplicationErrorCode.VALIDATION_ERROR,
                    "Контракт и реализация сервиса должны быть указаны"
            );
        }

        services.put(contract, contract.cast(implementation));
    }

    /**
     * Resolves a registered implementation by its contract.
     *
     * @param <T> service contract type
     * @param contract requested service contract
     * @return registered implementation
     * @throws ApplicationException when the contract is not registered
     */
    public <T> T resolve(final Class<T> contract) {
        if (contract == null) {
            throw new ApplicationException(
                    ApplicationErrorCode.VALIDATION_ERROR,
                    "Контракт запрашиваемого сервиса должен быть указан"
            );
        }

        var implementation = services.get(contract);
        if (implementation == null) {
            throw new ApplicationException(
                    ApplicationErrorCode.SERVICE_NOT_FOUND,
                    "Сервис не зарегистрирован: " + contract.getName()
            );
        }

        return contract.cast(implementation);
    }
}
