package pro290.vaporgame.PRO290VaporGameAPI.Controller;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.HttpStatus;
import org.springframework.web.bind.annotation.*;
import pro290.vaporgame.PRO290VaporGameAPI.Model.Game;

import java.util.Optional;

@RestController
@RequestMapping(value = "/game")
public class GameRestController {

    @Autowired
    private GameRepository repository;

    @GetMapping("/Example")
    @ResponseStatus(HttpStatus.OK)
    public Game GetExampleGame() {
        return new Game("League Of Legends",
                "A 2009 multiplayer online battle arena video game developed and published by Riot Games. Inspired by Defense of the Ancients, a custom map for Warcraft III, Riot's founders sought to develop a stand-alone game in the same genre.",
                    "Riot Games");
    }

    @GetMapping("/{id}")
    @ResponseStatus(HttpStatus.OK)
    public Optional<Game> GetGames(@PathVariable String id) {
        return repository.findById(id);
    }

    @PostMapping("/")
    @ResponseStatus(HttpStatus.CREATED)
    public String CreateGames(Game game) {
        return repository.save(game).getId();
    }

    @PatchMapping("/{id}")
    @ResponseStatus(HttpStatus.OK)
    public Game UpdateGames(@PathVariable String id, Game game) {
        Game ogGame = repository.findById(id).get();
        game.setId(ogGame.getId());
        game.setCreationDate(ogGame.getCreationDate());

        return repository.save(game);
    }

    @DeleteMapping("/{id}")
    @ResponseStatus(HttpStatus.OK)
    public void DeleteGames(@PathVariable String id) {
        repository.deleteById(id);
    }






}
