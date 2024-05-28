using Microsoft.AspNetCore.Mvc;
using Microsoft.EntityFrameworkCore;
using AutoMapper;
using Microsoft.AspNetCore.Authentication.JwtBearer;
using Microsoft.AspNetCore.Authorization;
using Microsoft.IdentityModel.Tokens;
using System.IdentityModel.Tokens.Jwt;
using System.Security.Claims;
using System.Text;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.Logging;

[ApiController]
[Route("[controller]")]
public class AuthController : ControllerBase
{
    private readonly ILogger<AuthController> _logger;

    private readonly IMapper _mapper;
    private readonly UserServiceDBContext _db;
    private readonly IConfiguration _config;

    public AuthController(ILogger<AuthController> logger, UserServiceDBContext db, IMapper mapper, IConfiguration config)
    {
        _logger = logger;
        _db = db;
        _mapper = mapper;
        _config = config;
    }

    [HttpGet]
    [Route("Test1")]
    public ActionResult<string> Test1()
    {
        return Ok("hello from AuthController");
    }

    [HttpPost]
    [Route("CreateTokenMethod1")]
    public async Task<ActionResult<string>> CreateTokenMethod1(UserDTO userDTO)
    {
        User? user = await _db.Users.Include(u => u.Orders).FirstOrDefaultAsync(u => u.Email == userDTO.Email && u.Password == userDTO.Password);

        if (user != null)
        {
            var authClaims = new List<Claim>
            {
                new Claim(ClaimTypes.Name, user.Username),
                new Claim(ClaimTypes.Email, user.Email),
                new Claim("UserGuid", user.UserGuid.ToString())
            };

            var authSigningKey = new SymmetricSecurityKey(Encoding.UTF8.GetBytes(_config["Jwt:Key"]));
            var token = new JwtSecurityToken(
                issuer: _config["Jwt:Issuer"],
                audience: _config["Jwt:Audience"],
                expires: DateTime.Now.AddHours(3),
                claims: authClaims,
                signingCredentials: new SigningCredentials(authSigningKey, SecurityAlgorithms.HmacSha256)
            );

            string finalToken = new JwtSecurityTokenHandler().WriteToken(token);
            return Ok(finalToken);
        }

        return Unauthorized();
    }

    [HttpPost]
    [Route("CreateTokenMethod2")]
    public async Task<IResult> CreateTokenMethod2(UserDTO userDTO)
    {
        User? user = await _db.Users.Include(u => u.Orders).FirstOrDefaultAsync(u => u.Email == userDTO.Email && u.Password == userDTO.Password);

        if (user != null)
        {
            var issuer = _config["Jwt:Issuer"];
            var audience = _config["Jwt:Audience"];
            var key = Encoding.ASCII.GetBytes(_config["Jwt:Key"]);
            var tokenDescriptor = new SecurityTokenDescriptor
            {
                Subject = new ClaimsIdentity(new[]
                {
                    new Claim("Id", Guid.NewGuid().ToString()),
                    new Claim(JwtRegisteredClaimNames.Sub, user.Username),
                    new Claim(JwtRegisteredClaimNames.Email, user.Email),
                    new Claim(JwtRegisteredClaimNames.Jti, Guid.NewGuid().ToString())
                }),
                Expires = DateTime.UtcNow.AddMinutes(5),
                Issuer = issuer,
                Audience = audience,
                SigningCredentials = new SigningCredentials(new SymmetricSecurityKey(key), SecurityAlgorithms.HmacSha512Signature)
            };

            var tokenHandler = new JwtSecurityTokenHandler();
            var token = tokenHandler.CreateToken(tokenDescriptor);
            var stringToken = tokenHandler.WriteToken(token);
            return Results.Ok(stringToken);
        }

        return Results.Unauthorized();
    }
}
