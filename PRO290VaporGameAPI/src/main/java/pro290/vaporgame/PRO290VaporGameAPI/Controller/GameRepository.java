package pro290.vaporgame.PRO290VaporGameAPI.Controller;

import org.socialsignin.spring.data.dynamodb.repository.EnableScan;
import org.springframework.data.repository.CrudRepository;
import pro290.vaporgame.PRO290VaporGameAPI.Model.Game;

@EnableScan
public interface GameRepository extends CrudRepository<Game, String> {

}
