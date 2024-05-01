package pro290.vaporgame.PRO290VaporGameAPI.Controller;

import org.springframework.http.HttpStatus;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.ResponseStatus;
import org.springframework.web.bind.annotation.RestController;
import pro290.vaporgame.PRO290VaporGameAPI.Model.Game;

@RestController
@RequestMapping(value = "/game")
public class GameRestController {

    @GetMapping("/")
    @ResponseStatus(HttpStatus.OK)
    public Game GetExampleGame() {
        return new Game("League Of Legends",
                "A 2009 multiplayer online battle arena video game developed and published by Riot Games. Inspired by Defense of the Ancients, a custom map for Warcraft III, Riot's founders sought to develop a stand-alone game in the same genre.",
                    "Riot Games");
    }
}
