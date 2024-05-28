using Microsoft.EntityFrameworkCore;
 
 
public class UserServiceDBContext : DbContext
{
    public UserServiceDBContext()
    {
    }
 
    public UserServiceDBContext(DbContextOptions<UserServiceDBContext> options) : base(options)
    {
    }
 

    protected override void OnConfiguring(DbContextOptionsBuilder optionsBuilder)
    {
    }
       public DbSet<User> Users { get; set; }
        public DbSet<Role> Roles { get; set; }
        public DbSet<UserRole> UserRoles { get; set; }
        public DbSet<Cart> Carts { get; set; }
        public DbSet<Library> Libraries { get; set; }
        public DbSet<LibraryGame> LibraryGames { get; set; }
        public DbSet<Game> Games { get; set; }
        public DbSet<Order> Orders { get; set; }

        protected override void OnModelCreating(ModelBuilder modelBuilder)
        {
            base.OnModelCreating(modelBuilder);

         // Define composite keys
        modelBuilder.Entity<UserRole>()
            .HasKey(ur => new { ur.UserID, ur.RoleID });

        modelBuilder.Entity<LibraryGame>()
            .HasKey(lg => new { lg.LibraryID, lg.GameID });

        modelBuilder.Entity<OrderGame>()
            .HasKey(og => new { og.OrderID, og.GameID });

        // Configure relationships

        // If User and Order are related and you need a UserGuid property in Order
        modelBuilder.Entity<Order>()
            .HasOne<User>() // Specify the type of the related entity
            .WithMany() // Indicate the relationship type
            .HasForeignKey(o => o.UserGuid) // Specify the foreign key property in the Order entity
            .OnDelete(DeleteBehavior.Restrict); // Define delete behavior

        // If Game and Order are related via OrderGuid
        modelBuilder.Entity<Game>()
            .HasOne<Order>() // Specify the type of the related entity
            .WithMany() // Indicate the relationship type
            .HasForeignKey(g => g.OrderGuid) // Specify the foreign key property in the Game entity
            .OnDelete(DeleteBehavior.Restrict); // Define delete behavior
        }
    }
