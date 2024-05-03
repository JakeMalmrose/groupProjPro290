package pro290.vaporgame.PRO290VaporGameAPI;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.context.annotation.ComponentScan;
import org.springframework.context.annotation.FilterType;
import org.springframework.data.jpa.repository.config.EnableJpaRepositories;
import pro290.vaporgame.PRO290VaporGameAPI.Controller.GameRepository;

@SpringBootApplication
@EnableJpaRepositories(excludeFilters =
@ComponentScan.Filter(type = FilterType.ASSIGNABLE_TYPE, value = GameRepository.class))
public class Pro290VaporGameApiApplication {

	public static void main(String[] args) {
		SpringApplication.run(Pro290VaporGameApiApplication.class, args);
	}

}
