using System.ComponentModel.DataAnnotations;


    public class LoginModel
    {
        [Required]
        public string Email { get; set; } = string.Empty;

        [Required]
        [DataType(DataType.Password)]
        public string Password { get; set; } = string.Empty;
    }
